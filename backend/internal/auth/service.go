package auth

//go:generate moq -out mock_oidc_client_test.go -pkg auth_test ../core/oidc GoogleClient
//go:generate moq -out mock_jwt_service_test.go -pkg auth_test ../core/jwt_d Service
//go:generate moq -out mock_youtube_client_test.go -pkg auth_test ../core/youtube_d Client:YouTubeClientMock
//go:generate moq -out mock_jti_blacklist_repository_test.go -pkg auth_test ../core/database_d JtiBlacklistRepository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
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
	ErrInvalidCSRFOrState     = core.NewDomainError("auth.invalid_csrf_or_state", "invalid csrf or state", core.StatusBadRequest)
	ErrInvalidCSRF            = core.NewDomainError("auth.invalid_csrf", "invalid csrf: csrf != state", core.StatusBadRequest)
	ErrIDTokenNotFound        = oidc.ErrIDTokenNotFound
	ErrNoImportOptionSelected = core.NewDomainError("auth.no_import_option_selected", "at least one import option must be selected", core.StatusBadRequest)
)

type Service struct {
	db *pgxpool.Pool

	oidcClient      oidc.GoogleClient
	channelService  *channel.Service
	playlistService *playlist.Service
	youtubeClient   youtube_d.Client

	jwtService       jwt_d.Service
	jtiBlacklistRepo database_d.JtiBlacklistRepository

	serverURL            string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration

	refreshTokenQS  RefreshTokenQueryService
	authorizationQS AuthorizationQueryService
	userQS          user.UserQueryService
}

func NewService(
	db *pgxpool.Pool,
	oidcClient oidc.GoogleClient,
	youtubeClient youtube_d.Client,
	channelService *channel.Service,
	playlistService *playlist.Service,
	serverURL string,
	jwtService jwt_d.Service,
	accessTokenDuration time.Duration,
	refreshTokenDuration time.Duration,
	jtiBlacklistRepo database_d.JtiBlacklistRepository,
) *Service {
	return new(Service{
		db:                   db,
		oidcClient:           oidcClient,
		youtubeClient:        youtubeClient,
		jwtService:           jwtService,
		jtiBlacklistRepo:     jtiBlacklistRepo,
		serverURL:            serverURL,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		refreshTokenQS:       NewRefreshTokenQueryService(db),
		authorizationQS:      NewAuthorizationQueryService(db),
		userQS:               user.NewUserQueryService(db),
		channelService:       channelService,
		playlistService:      playlistService,
	})
}

func (s *Service) CreateAuthCode(ctx context.Context, platform string) (_, _ string, err error) {
	defer util.Wrap(&err, "auth.(*Service).CreateAuthCode")

	if platform == "" {
		platform = "web"
	}

	stateToken, err := s.jwtService.SignOIDCStateToken(platform, s.serverURL)
	if err != nil {
		return "", "", err
	}

	return s.oidcClient.AuthCodeURL(stateToken), stateToken, nil
}

type GoogleOIDCCallbackResult struct {
	AccessToken           string
	RefreshToken          string
	CSRFToken             string
	RedirectPath          string
	Platform              string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, csrf, state, code, ipAddress, countryCode, userAgent string) (_ GoogleOIDCCallbackResult, err error) {
	defer util.Wrap(&err, "auth.(*Service).GoogleOIDCCallback")

	if csrf == "" || state == "" {
		return GoogleOIDCCallbackResult{}, ErrInvalidCSRFOrState
	}
	if csrf != state {
		return GoogleOIDCCallbackResult{}, ErrInvalidCSRF
	}

	platform, err := s.jwtService.VerifyOIDCStateToken(state)
	if err != nil {
		return GoogleOIDCCallbackResult{}, ErrInvalidCSRFOrState
	}

	sub, err := s.oidcClient.ExchangeAndVerify(ctx, code)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	authorization, err := NewAuthorization("https://accounts.google.com", sub) // TODO: DI
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}

	authorizationID, err := NewAuthorizationRepository(q).Save(ctx, authorization)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}

	// もし、userテーブルに存在するなら、ログイン用リフレッシュトークンを作成してダッシュボードにリダイレクトさせる。
	// そうでないなら、登録用リフレッシュトークンを作成して、ユーザー登録にリダイレクトさせる。
	// リフレッシュトークンは、user_authorizationに紐づくものである。
	// どちらにせよ、リフレッシュトークンは発行する。
	refreshTokenRawStr, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}
	refreshToken, err := NewRefreshToken(
		userAgent,
		ipAddress,
		countryCode,
		"", // TODO: cityName
		time.Now().UTC().Add(s.refreshTokenDuration),
		WithRefreshTokenRaw(refreshTokenRawStr),
	)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}
	_, err = NewRefreshTokenRepository(q).Save(ctx, authorizationID, refreshToken)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}

	// csrfはどのみち必要になるのでここで作っておく
	csrfGenerated, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return GoogleOIDCCallbackResult{}, err
	}

	// userテーブルに存在する場合、リフレッシュトークンを保存して、アクセストークンを発行する。
	// userテーブルに存在しない場合、登録用アクセストークンを発行する。

	// userテーブルに存在するかどうか
	userPublicID, isDeactivated, err := s.userQS.FindByAuthorizationID(ctx, authorizationID)

	if err == nil && !isDeactivated { // 現役で存在する場合
		at, atExp, err := s.jwtService.SignUserAccessToken(userPublicID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		return GoogleOIDCCallbackResult{
			AccessToken:           at,
			RefreshToken:          refreshTokenRawStr,
			CSRFToken:             csrfGenerated,
			RedirectPath:          "",
			Platform:              platform,
			AccessTokenExpiresAt:  atExp,
			RefreshTokenExpiresAt: refreshToken.ExpiresAt,
		}, nil
	} else if err == nil && isDeactivated { // 退会済みだが、レコードが残っている場合
		accessToken, atExp, err := s.jwtService.SignRegisterToken(authorization.ID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		return GoogleOIDCCallbackResult{
			AccessToken:           accessToken,
			RefreshToken:          refreshTokenRawStr,
			CSRFToken:             csrfGenerated,
			RedirectPath:          "reactivation",
			Platform:              platform,
			AccessTokenExpiresAt:  atExp,
			RefreshTokenExpiresAt: refreshToken.ExpiresAt,
		}, nil
	} else if errors.Is(err, core.ErrNotFound) { // 存在しない場合
		accessToken, atExp, err := s.jwtService.SignRegisterToken(authorization.ID, refreshToken.AccessTokenJTI, s.serverURL)
		if err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		if err := tx.Commit(ctx); err != nil {
			return GoogleOIDCCallbackResult{}, err
		}

		return GoogleOIDCCallbackResult{
			AccessToken:           accessToken,
			RefreshToken:          refreshTokenRawStr,
			CSRFToken:             csrfGenerated,
			RedirectPath:          "register",
			Platform:              platform,
			AccessTokenExpiresAt:  atExp,
			RefreshTokenExpiresAt: refreshToken.ExpiresAt,
		}, nil
	} else { // ただのDBエラー
		return GoogleOIDCCallbackResult{}, err
	}
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) (err error) {
	defer util.Wrap(&err, "auth.(*Service).Logout")

	q := sqlc.New(s.db)
	userID, jti, jtiExpiresAt, err := s.jwtService.VerifyUserAccessToken(accessToken)
	if err != nil {
		return err
	}

	refreshTokenHash := util.Sha256Hex(refreshToken)
	if err := NewRefreshTokenRepository(q).RevokeByTokenHash(ctx, userID, refreshTokenHash); err != nil {
		return err
	}

	return s.jtiBlacklistRepo.InsertJTI(ctx, jti, jtiExpiresAt)
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, userAgent string) (_, _ string, _, _ time.Time, err error) {
	defer util.Wrap(&err, "auth.(*Service).RefreshToken")

	tokenHash := util.Sha256Hex(refreshToken)

	newToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	newRefreshToken, err := NewRefreshToken(
		userAgent,
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

	userID, err := NewRefreshTokenRepository(q).RotateRefreshToken(ctx, newRefreshToken, tokenHash, time.Now().UTC().Add(-s.accessTokenDuration).UTC())
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
	defer util.Wrap(&err, "auth.(*Service).GetSessions")

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
	defer util.Wrap(&err, "auth.(*Service).RemoveSession")

	q := sqlc.New(s.db)
	removedPublicID, accessTokenJTI, err := NewRefreshTokenRepository(q).RevokeByID(ctx, userID, sessionID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.jtiBlacklistRepo.InsertJTI(ctx, accessTokenJTI, time.Now().UTC().Add(s.accessTokenDuration)); err != nil {
		return uuid.Nil, err
	}

	return removedPublicID, nil
}

func (s *Service) CreateYouTubeAuthCode(_ context.Context, userID uuid.UUID, importSubscriptions, importLikes bool) (_ string, err error) {
	defer util.Wrap(&err, "auth.(*Service).CreateYouTubeAuthCode")

	if !importSubscriptions && !importLikes {
		return "", ErrNoImportOptionSelected
	}

	state, err := s.jwtService.SignYouTubeImportStateToken(userID, importSubscriptions, importLikes, s.serverURL)
	if err != nil {
		return "", err
	}

	return s.youtubeClient.OAuthAuthCodeURL(state), nil
}

func (s *Service) YouTubeOAuthCallback(ctx context.Context, state, code string) (err error) {
	defer util.Wrap(&err, "auth.(*Service).YouTubeOAuthCallback")

	userID, importSubscriptions, importLikes, err := s.jwtService.VerifyYouTubeImportStateToken(state)
	if err != nil {
		return ErrInvalidCSRFOrState
	}

	oauthClient, err := s.youtubeClient.OAuthExchange(ctx, code)
	if err != nil {
		return err
	}

	if importSubscriptions {
		channels, err := oauthClient.FetchAllSubscriptions(ctx)
		if err != nil {
			return err
		}
		for _, channel := range channels {
			if _, err := s.channelService.SubscribeChannel(ctx, userID, string(channel.ID)); err != nil {
				util.LoggerFromContext(ctx).InfoContext(ctx, "failed to subscribe channel(youtube oauth callback)", slog.String("channel_id", string(channel.ID)))
			}
		}
		clear(channels)
	}

	if importLikes {
		if _, err := s.playlistService.CreatePlaylistWithOAuthClient(ctx, userID, "高評価した動画", "", "private", "normal", oauthClient, "LL"); err != nil {
			util.LoggerFromContext(ctx).InfoContext(ctx, "failed to import liked playlist(youtube oauth callback)", slog.Any("error", err))
		}
	}

	return nil
}

type ReactivateAccountResult struct {
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

func (s *Service) ReactivateAccount(ctx context.Context, registerAccessToken, ipAddress, countryCode, userAgent string) (_ ReactivateAccountResult, err error) {
	defer util.Wrap(&err, "auth.(*Service).ReactivateAccount")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ReactivateAccountResult{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	// register tokenの検証
	authorizationPublicID, jti, err := s.jwtService.VerifyRegisterToken(registerAccessToken)
	if err != nil {
		return ReactivateAccountResult{}, err
	}
	// jti blacklist検証
	blacklisted, err := s.jtiBlacklistRepo.IsJtiExist(ctx, jti)
	if err != nil {
		return ReactivateAccountResult{}, err
	}
	if blacklisted {
		return ReactivateAccountResult{}, core.ErrJTIBlacklisted
	}

	// authorizationIDで勧告ロック
	if err := database_d.TryAdLock(ctx, q, authorizationPublicID[:]); err != nil {
		return ReactivateAccountResult{}, err
	}

	// h_userからm_userに行を戻す。元のm_user_idを保つので関連データのFKがそのまま生きる。
	authorizationID, userPublicID, err := NewAuthorizationRepository(q).RestoreUserFromHistory(ctx, authorizationPublicID)
	if err != nil {
		return ReactivateAccountResult{}, err
	}

	// 復活後のセッション用にrefresh tokenを新規発行する
	refreshTokenRawStr, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return ReactivateAccountResult{}, err
	}
	refreshToken, err := NewRefreshToken(
		userAgent,
		ipAddress,
		countryCode,
		"", // TODO: cityName
		time.Now().UTC().Add(s.refreshTokenDuration),
		WithRefreshTokenRaw(refreshTokenRawStr),
	)
	if err != nil {
		return ReactivateAccountResult{}, err
	}
	if _, err := NewRefreshTokenRepository(q).Save(ctx, authorizationID, refreshToken); err != nil {
		return ReactivateAccountResult{}, err
	}

	accessToken, accessTokenExpiresAt, err := s.jwtService.SignUserAccessToken(userPublicID, refreshToken.AccessTokenJTI, s.serverURL)
	if err != nil {
		return ReactivateAccountResult{}, err
	}

	// 使用済みregisterトークンのJTIをブラックリストに追加
	if err := s.jtiBlacklistRepo.InsertJTI(ctx, jti, time.Now().UTC().Add(s.accessTokenDuration)); err != nil {
		return ReactivateAccountResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ReactivateAccountResult{}, err
	}

	return ReactivateAccountResult{
		AccessToken:           accessToken,
		RefreshToken:          refreshTokenRawStr,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshToken.ExpiresAt,
	}, nil
}
