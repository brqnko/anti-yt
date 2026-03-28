package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/playlist"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCSRFOrState = core.NewDomainError("auth.invalid_csrf_or_state", "invalid csrf or state")
	ErrInvalidCSRF        = core.NewDomainError("auth.invalid_csrf", "invalid csrf: csrf != state")
	ErrIDTokenNotFound    = oidc.ErrIDTokenNotFound
)

type Service struct {
	db *pgxpool.Pool

	oidcService     oidc.GoogleOIDCService
	channelService  *channel.Service
	playlistService *playlist.Service
	ytService       youtube_d.Service

	jwtService jwt_d.Service

	serverURL            string
	refreshTokenDuration time.Duration

	refreshTokenQS  RefreshTokenQueryService
	authorizationQS AuthorizationQueryService
	userQS          user.UserQueryService
}

func NewService(
	db *pgxpool.Pool,
	oidcService oidc.GoogleOIDCService,
	youtubeService youtube_d.Service,
	channelService *channel.Service,
	playlistService *playlist.Service,
	serverURL string,
	jwtService jwt_d.Service,
	refreshTokenDuration time.Duration,
) *Service {
	return &Service{
		db:                   db,
		oidcService:          oidcService,
		ytService:            youtubeService,
		jwtService:           jwtService,
		serverURL:            serverURL,
		refreshTokenDuration: refreshTokenDuration,
		refreshTokenQS:       NewRefreshTokenQueryService(db),
		authorizationQS:      NewAuthorizationQueryService(db),
		userQS:               user.NewUserQueryService(db),
		channelService:       channelService,
		playlistService:      playlistService,
	}
}

func (s *Service) CreateAuthCode(ctx context.Context) (_, _ string, err error) {
	defer util.Wrap(&err, "Service.CreateAuthCode")

	csrfToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", err
	}

	return s.oidcService.AuthCodeURL(csrfToken), csrfToken, nil
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, csrf, state, code, ipAddress, countryCode, deviceFingerprint, userAgent string) (_, _, _, _ string, _, _ time.Time, err error) {
	defer util.Wrap(&err, "Service.GoogleOIDCCallback")

	if csrf == "" || state == "" {
		return "", "", "", "", time.Time{}, time.Time{}, ErrInvalidCSRFOrState
	}
	if csrf != state {
		return "", "", "", "", time.Time{}, time.Time{}, ErrInvalidCSRF
	}

	sub, err := s.oidcService.ExchangeAndVerify(ctx, code)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	authorization, err := NewAuthorization("https://accounts.google.com", sub) // TODO: DI
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}

	authorizationID, err := NewAuthorizationRepository(q).Save(ctx, authorization)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}

	// もし、userテーブルに存在するなら、ログイン用リフレッシュトークンを作成してダッシュボードにリダイレクトさせる。
	// そうでないなら、登録用リフレッシュトークンを作成して、ユーザー登録にリダイレクトさせる。
	// リフレッシュトークンは、user_authorizationに紐づくものである。
	// どちらにせよ、リフレッシュトークンは発行する。
	refreshTokenRawStr, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	refreshToken, err := NewRefreshToken(
		userAgent,
		deviceFingerprint,
		ipAddress,
		countryCode,
		"", // TODO: cityName
		time.Now().UTC().Add(s.refreshTokenDuration),
		WithRefreshTokenRaw(refreshTokenRawStr),
	)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
	_, err = NewRefreshTokenRepository(q).Save(ctx, authorizationID, refreshToken)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}

	// csrfはどのみち必要になるのでここで作っておく
	csrfGenerated, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", "", "", time.Time{}, time.Time{}, err
	}

	// userテーブルに存在する場合、リフレッシュトークンを保存して、アクセストークンを発行する。
	// userテーブルに存在しない場合、登録用アクセストークンを発行する。

	// userテーブルに存在するかどうか
	userPublicID, isDeactivated, err := s.userQS.FindByAuthorizationID(ctx, authorizationID)

	if err == nil && !isDeactivated { // 現役で存在する場合
		at, atExp, err := s.jwtService.SignUserAccessToken(userPublicID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		return at, refreshTokenRawStr, csrfGenerated, "dashboard", atExp, refreshToken.ExpiresAt, nil
	} else if err == nil && isDeactivated { // 退会済みだが、レコードが残っている場合
		at, atExp, err := s.jwtService.SignRegisterToken(authorization.ID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		return at, refreshTokenRawStr, csrfGenerated, "reactivation", atExp, refreshToken.ExpiresAt, nil
	} else if errors.Is(err, pgx.ErrNoRows) { // 存在しない場合
		at, atExp, err := s.jwtService.SignRegisterToken(authorization.ID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return "", "", "", "", time.Time{}, time.Time{}, err
		}

		return at, refreshTokenRawStr, csrfGenerated, "register", atExp, refreshToken.ExpiresAt, nil
	} else { // ただのDBエラー
		return "", "", "", "", time.Time{}, time.Time{}, err
	}
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) (err error) {
	defer util.Wrap(&err, "Service.Logout")

	q := sqlc.New(s.db)
	userID, _, _, _ := s.jwtService.VerifyUserAccessToken(accessToken)

	refreshTokenHash := util.Sha256Hex(refreshToken)
	return NewRefreshTokenRepository(q).RevokeByTokenHash(ctx, userID, refreshTokenHash, time.Now().UTC().Add(s.refreshTokenDuration))
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, deviceFingerprint, userAgent string) (_, _ string, _, _ time.Time, err error) {
	defer util.Wrap(&err, "Service.RefreshToken")

	tokenHash := util.Sha256Hex(refreshToken)

	newToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	newRefreshToken, err := NewRefreshToken(
		userAgent,
		deviceFingerprint,
		ipAddress,
		countryCode,
		"",
		time.Now().UTC().Add(s.refreshTokenDuration),
		WithRefreshTokenRaw(newToken),
	)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	userID, err := NewRefreshTokenRepository(q).RotateRefreshToken(ctx, newRefreshToken, tokenHash, time.Now().UTC().Add(-s.jwtService.TokenDuration()).UTC())
	if err != nil { // 条件を満たさない、あるいはDBエラー
		return "", "", time.Time{}, time.Time{}, err
	}

	accessTokenString, signedAtExpiresAtD, err := s.jwtService.SignUserAccessToken(userID, newRefreshToken.AccessTokenJTI, s.serverURL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return newToken, accessTokenString, signedAtExpiresAtD, newRefreshToken.ExpiresAt, nil
}

func (s *Service) GetSessions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (_ []GetSessionsView, _ bool, err error) {
	defer util.Wrap(&err, "Service.GetSessions")

	view, err := s.refreshTokenQS.GetSessions(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}
	if len(view) > int(limit) {
		return view[:limit], true, nil
	}
	return view, false, nil
}

func (s *Service) RemoveSession(ctx context.Context, userID, sessionID uuid.UUID) (_ uuid.UUID, err error) {
	defer util.Wrap(&err, "Service.RemoveSession")

	q := sqlc.New(s.db)
	removedPublicID, err := NewRefreshTokenRepository(q).RevokeByID(ctx, userID, sessionID, time.Now().UTC().Add(s.jwtService.TokenDuration()))
	if err != nil {
		return uuid.Nil, err
	}

	return removedPublicID, nil
}

func (s *Service) CreateYouTubeAuthCode(_ context.Context, userID uuid.UUID) (_ string, err error) {
	defer util.Wrap(&err, "Service.CreateYouTubeAuthCode")

	state, err := s.jwtService.SignOAuthStateToken(userID, "youtube", s.serverURL)
	if err != nil {
		return "", err
	}

	return s.ytService.OAuthAuthCodeURL(state), nil
}

func (s *Service) YouTubeOAuthCallback(ctx context.Context, state, code string) (err error) {
	defer util.Wrap(&err, "Service.YouTubeOAuthCallback")

	userID, provider, err := s.jwtService.VerifyOAuthStateToken(state)
	if err != nil || provider != "youtube" {
		return ErrInvalidCSRFOrState
	}

	ytAccessToken, err := s.ytService.OAuthExchange(ctx, code)
	if err != nil {
		return err
	}

	// 登録してるチャンネルを取得
	channels, err := s.ytService.FetchAllSubscriptions(ctx, ytAccessToken)
	if err != nil {
		return err
	}
	// 登録
	for _, channel := range channels {
		if _, err := s.channelService.SubscribeChannel(ctx, userID, string(channel.ID)); err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to subscribe channel(youtube oauth callback)", slog.String("channel_id", string(channel.ID)))
		}
	}
	clear(channels)

	// 履歴のimport
	// NOTE: YouTube Data API v3 は Watch History (HL) へのアクセスをサポートしていないためコメントアウト
	// historyPageToken := ""
	// allHistory := make([]youtube_d.WatchHistory, 0)
	// for {
	// 	histories, nextPageToken, err := s.ytService.FetchWatchHistory(ctx, ytAccessToken, historyPageToken)
	// 	if err != nil {
	// 		slog.Info("failed to fetch watch history(youtube oauth callback)", "error", err)
	// 		break
	// 	}
	// 	allHistory = append(allHistory, histories...)
	// 	if nextPageToken == "" {
	// 		break
	// 	}
	// 	historyPageToken = nextPageToken
	// }
	// historyVideos := make(map[youtube_d.VideoID]youtube_d.Video)
	// videoIDsToRequest := make([]youtube_d.VideoID, 0, 50)
	// fetchVideos := func() error {
	// 	if len(videoIDsToRequest) == 0 {
	// 		return nil
	// 	}
	// 	videos, err := s.ytService.FetchVideoDetail(ctx, videoIDsToRequest)
	// 	if err != nil {
	// 		slog.Info("failed to fetch video detail(youtube oauth callback)", "error", err)
	// 		return err
	// 	}
	// 	for id, v := range videos {
	// 		historyVideos[id] = v
	// 	}
	// 	videoIDsToRequest = videoIDsToRequest[:0]
	// 	return nil
	// }
	// for _, h := range allHistory {
	// 	if _, ok := historyVideos[h.VideoID]; ok {
	// 		continue
	// 	}
	// 	videoIDsToRequest = append(videoIDsToRequest, h.VideoID)
	// 	if len(videoIDsToRequest) >= 50 {
	// 		if err := fetchVideos(); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// if err := fetchVideos(); err != nil {
	// 	return err
	// }

	// // 履歴のチャンネル情報を取得
	// historyChannels := make(map[youtube_d.ChannelID]youtube_d.Channel)
	// channelIDsToRequest := make([]youtube_d.ChannelID, 0, 50)
	// fetchChannels := func() error {
	// 	if len(channelIDsToRequest) == 0 {
	// 		return nil
	// 	}
	// 	channels, err := s.ytService.FetchChannelDetail(ctx, channelIDsToRequest)
	// 	if err != nil {
	// 		slog.Info("failed to fetch channel detail(youtube oauth callback)", "error", err)
	// 		return err
	// 	}
	// 	for id, c := range channels {
	// 		historyChannels[id] = c
	// 	}
	// 	channelIDsToRequest = channelIDsToRequest[:0]
	// 	return nil
	// }
	// for _, v := range historyVideos {
	// 	if _, ok := historyChannels[v.ChannelID]; ok {
	// 		continue
	// 	}
	// 	channelIDsToRequest = append(channelIDsToRequest, v.ChannelID)
	// 	if len(channelIDsToRequest) >= 50 {
	// 		if err := fetchChannels(); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	// if err := fetchChannels(); err != nil {
	// 	return err
	// }

	// // チャンネル情報をinsert
	// fetchedAt := time.Now().UTC()
	// channelUUIDs := make(map[youtube_d.ChannelID]uuid.UUID)
	// for ytChannelID, c := range historyChannels {
	// 	ch, err := channel.NewChannel(fetchedAt, fetchedAt, c)
	// 	if err != nil {
	// 		slog.Info("failed to new channel(youtube oauth callback)", "error", err)
	// 		continue
	// 	}
	// 	if _, err := channel.NewChannelRepository(sqlc.New(s.db)).Save(ctx, ch); err != nil {
	// 		slog.Info("failed to save channel(youtube oauth callback)", "error", err)
	// 		continue
	// 	}
	// 	channelUUIDs[ytChannelID] = ch.ID
	// }

	// // 動画情報をinsert
	// videoUUIDs := make(map[youtube_d.VideoID]uuid.UUID)
	// for ytVideoID, v := range historyVideos {
	// 	channelUUID, ok := channelUUIDs[v.ChannelID]
	// 	if !ok {
	// 		slog.Info("channel uuid not found(youtube oauth callback)", "channel_id", v.ChannelID)
	// 		continue
	// 	}
	// 	vd, err := video.NewVideo(channelUUID, fetchedAt, v)
	// 	if err != nil {
	// 		slog.Info("failed to new video(youtube oauth callback)", "error", err)
	// 		continue
	// 	}
	// 	if _, err := video.NewVideoRepository(sqlc.New(s.db)).Save(ctx, vd); err != nil {
	// 		slog.Info("failed to save video(youtube oauth callback)", "error", err)
	// 		continue
	// 	}
	// 	videoUUIDs[ytVideoID] = vd.ID
	// }

	// // 再生履歴をinsert
	// for _, h := range allHistory {
	// 	videoUUID, ok := videoUUIDs[h.VideoID]
	// 	if !ok {
	// 		slog.Info("video uuid not found(youtube oauth callback)", "video_id", h.VideoID)
	// 		continue
	// 	}
	// 	v, ok := historyVideos[h.VideoID]
	// 	if !ok {
	// 		continue
	// 	}
	// 	watchStartAt := h.WatchedAt.Add(-time.Duration(v.LengthSeconds) * time.Second)
	// 	if err := history.NewHistoryRepository(sqlc.New(s.db)).Import(ctx, userID, videoUUID, watchStartAt, h.WatchedAt); err != nil {
	// 		slog.Info("failed to import history(youtube oauth callback)", "error", err)
	// 		continue
	// 	}
	// }
	// clear(allHistory)
	// clear(historyVideos)
	// clear(historyChannels)
	// clear(channelUUIDs)
	// clear(videoUUIDs)

	// 高評価プレイリストをimport
	if _, err := s.playlistService.CreatePlaylistWithAccessToken(ctx, userID, "高評価した動画", "", "private", "normal", ytAccessToken, "LL"); err != nil {
		util.LoggerFromContext(ctx).InfoContext(ctx, "failed to import liked playlist(youtube oauth callback)", slog.Any("error", err))
	}

	return nil
}
