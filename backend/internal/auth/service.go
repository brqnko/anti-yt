package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mssola/user_agent"
)

var (
	ErrInvalidCSRFOrState = errors.New("invalid csrf or state")
	ErrInvalidCSRF        = errors.New("invalid csrf: csrf != state")
	ErrIDTokenNotFound    = oidc.ErrIDTokenNotFound

	ErrNoSuchRefreshToken = errors.New("no such refresh token")
)

type GoogleOIDCCallbackParams struct {
	CSRF              string
	State             string
	Code              string
	IPAddress         string
	CountryCode       string
	DeviceFingerprint string
	UserAgent         string
}

type GoogleOIDCCallbackResult struct {
	AccessToken           string
	RefreshToken          string
	CSRFToken             string
	RedirectPath          string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type Service struct {
	db *pgxpool.Pool

	oidcService oidc.GoogleOIDCService

	jwtService jwt_d.JWTService

	serverURL            string
	refreshTokenDuration time.Duration
}

func NewService(db *pgxpool.Pool, oidcService oidc.GoogleOIDCService, serverURL string, jwtService jwt_d.JWTService, refreshTokenDuration time.Duration) (*Service, error) {
	return &Service{
		db:                   db,
		oidcService:          oidcService,
		jwtService:           jwtService,
		serverURL:            serverURL,
		refreshTokenDuration: refreshTokenDuration,
	}, nil
}

func (s *Service) CreateAuthCode(ctx context.Context) (redirectURL, csrf string, err error) {
	csrfToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", err
	}

	return s.oidcService.AuthCodeURL(csrfToken), csrfToken, nil
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, params GoogleOIDCCallbackParams) (*GoogleOIDCCallbackResult, error) {
	if params.CSRF == "" || params.State == "" {
		return nil, ErrInvalidCSRFOrState
	}
	if params.CSRF != params.State {
		return nil, ErrInvalidCSRF
	}

	sub, err := s.oidcService.ExchangeAndVerify(ctx, params.Code)
	if err != nil {
		return nil, fmt.Errorf("ExchangeAndVerify: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	saveAuthorization, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer: "https://accounts.google.com", // TODO: DI
		Sub:    sub,
	})
	if err != nil {
		return nil, fmt.Errorf("SaveAuthorization: %w", err)
	}

	// もし、userテーブルに存在するなら、ログイン用リフレッシュトークンを作成してダッシュボードにリダイレクトさせる。
	// そうでないなら、登録用リフレッシュトークンを作成して、ユーザー登録にリダイレクトさせる。
	// リフレッシュトークンは、user_authorizationに紐づくものである。
	// どちらにせよ、リフレッシュトークンは発行する。
	refreshTokenRaw, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return nil, err
	}
	refreshTokenHashRaw := sha256.Sum256([]byte(refreshTokenRaw))
	refreshTokenHash := hex.EncodeToString(refreshTokenHashRaw[:]) // NOTE: URLセーフ
	refreshTokenExpiresAt := time.Now().UTC().Add(s.refreshTokenDuration)

	accessTokenJti, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	// リフレッシュトークンの保存
	ua := user_agent.New(params.UserAgent)
	browserName, browserVersion := ua.Browser()
	_, err = q.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
		MUserAuthorizationID: saveAuthorization.MUserAuthorizationID,
		TokenHash:            refreshTokenHash,
		IpAddress:            params.IPAddress,
		DeviceFingerprint:    params.DeviceFingerprint,
		UserAgent:            params.UserAgent,
		CountryCode:          params.CountryCode,
		CityName:             "", // TODO: 今はなし
		BrowserName:          fmt.Sprintf("%s:%s", browserName, browserVersion),
		DeviceType:           ua.OSInfo().FullName,
		ExpiresAt:            refreshTokenExpiresAt.UTC(),
		AccessTokenJti:       accessTokenJti,
	})
	if err != nil {
		return nil, fmt.Errorf("SaveRefreshToken: %w", err)
	}
	csrf, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return nil, err
	}

	// userテーブルに存在する場合、リフレッシュトークンを保存して、アクセストークンを発行する。
	// userテーブルに存在しない場合、登録用アクセストークンを発行する。

	// userテーブルに存在するかどうか
	getUserByAuthorization, err := q.GetUserIDByAuthorization(ctx, saveAuthorization.MUserAuthorizationID)

	if err == nil && !getUserByAuthorization.IsH { // 現役で存在する場合
		accessToken, accessTokenExpiresAt, err := s.jwtService.SignUserAccessToken(getUserByAuthorization.PublicID, accessTokenJti, s.serverURL)
		if err != nil {
			return nil, fmt.Errorf("SignUserAccessToken: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("Commit: %w", err)
		}

		return &GoogleOIDCCallbackResult{
			AccessToken:           accessToken,
			RefreshToken:          refreshTokenRaw,
			CSRFToken:             csrf,
			RedirectPath:          "dashboard",
			AccessTokenExpiresAt:  accessTokenExpiresAt,
			RefreshTokenExpiresAt: refreshTokenExpiresAt,
		}, nil
	} else if err == nil && getUserByAuthorization.IsH { // 退会済みだが、レコードが残っている場合
		accessToken, accessTokenJtiExpiresAt, err := s.jwtService.SignRegisterToken(saveAuthorization.PublicID, accessTokenJti, s.serverURL)
		if err != nil {
			return nil, fmt.Errorf("SignRegisterToken: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("Commit: %w", err)
		}

		return &GoogleOIDCCallbackResult{
			AccessToken:           accessToken,
			RefreshToken:          refreshTokenRaw,
			CSRFToken:             csrf,
			RedirectPath:          "reactivation",
			AccessTokenExpiresAt:  accessTokenJtiExpiresAt,
			RefreshTokenExpiresAt: refreshTokenExpiresAt,
		}, nil
	} else if errors.Is(err, pgx.ErrNoRows) { // 存在しない場合
		accessToken, accessTokenJtiExpiresAt, err := s.jwtService.SignRegisterToken(saveAuthorization.PublicID, accessTokenJti, s.serverURL)
		if err != nil {
			return nil, fmt.Errorf("SignRegisterToken: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("Commit: %w", err)
		}

		return &GoogleOIDCCallbackResult{
			AccessToken:           accessToken,
			RefreshToken:          refreshTokenRaw,
			CSRFToken:             csrf,
			RedirectPath:          "register",
			AccessTokenExpiresAt:  accessTokenJtiExpiresAt,
			RefreshTokenExpiresAt: refreshTokenExpiresAt,
		}, nil
	} else { // ただのDBエラー
		return nil, fmt.Errorf("GetUserIDByAuthorization: %w", err)
	}
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	q := sqlc.New(s.db)
	userId, _, _, _ := s.jwtService.VerifyUserAccessToken(accessToken)

	refreshTokenHashRaw := sha256.Sum256([]byte(refreshToken))
	refreshTokenHash := hex.EncodeToString(refreshTokenHashRaw[:])
	if _, err := q.RemoveRefreshTokenByTokenHashAndSaveJtiBlacklist(ctx, sqlc.RemoveRefreshTokenByTokenHashAndSaveJtiBlacklistParams{
		UserPublicID: userId,
		TokenHash:    refreshTokenHash,
		ExpiresAt:    time.Now().UTC().Add(s.refreshTokenDuration),
	}); err != nil {
		return fmt.Errorf("RemoveRefreshTokenByTokenHashAndSaveJtiBlacklist: %w", err)
	}

	return nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, deviceFingerprint, userAgent string) (newRefreshToken, newAccessToken string, accessTokenExpiresAt, refreshTokenExpiresAt time.Time, err error) {
	newAccessTokenJti, err := uuid.NewV7()
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	tokenHashRaw := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(tokenHashRaw[:])

	newToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	newTokenHashRaw := sha256.Sum256([]byte(newToken))
	newTokenHash := hex.EncodeToString(newTokenHashRaw[:])

	now := time.Now().UTC()
	newTokenExpiresAt := now.Add(s.refreshTokenDuration).UTC()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("Begin: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	ua := user_agent.New(userAgent)
	browserName, browserVersion := ua.Browser()
	userPublicId, err := q.UpdateRefreshToken(ctx, sqlc.UpdateRefreshTokenParams{
		NewTokenHash:         newTokenHash,
		NewExpiresAt:         newTokenExpiresAt,
		NewIpAddress:         ipAddress,
		NewDeviceFingerprint: deviceFingerprint,
		NewUserAgent:         userAgent,
		NewCountryCode:       countryCode,
		NewCityName:          "",
		NewBrowserName:       fmt.Sprintf("%s:%s", browserName, browserVersion),
		NewDeviceType:        ua.OSInfo().FullName,
		TokenHashForCheck:    tokenHash, // token_hash = token_hash_for_check
		// updated_at < @updated_at_for_checkがsqlcの引数
		// updated_at + token_duration < now にしたい。
		// updated_at < now - token_duration 変形するとこうなる。
		// すなわち、@updated_at_for_check = now - token_duration
		UpdatedAtForCheck: now.Add(-s.jwtService.TokenDuration()).UTC(),
		NewAccessTokenJti: newAccessTokenJti,
	})
	if err != nil { // 条件を満たさない、あるいはDBエラー
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("UpdateRefreshToken: %w", err)
	}

	accessTokenString, signedAtExpiresAtD, err := s.jwtService.SignUserAccessToken(userPublicId, newAccessTokenJti, s.serverURL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("SignUserAccessToken: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", time.Time{}, time.Time{}, fmt.Errorf("Commit: %w", err)
	}

	return newToken, accessTokenString, signedAtExpiresAtD, newTokenExpiresAt, nil
}

func (s *Service) GetSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	q := sqlc.New(s.db)

	sessions, err := q.GetRefreshTokens(ctx, sqlc.GetRefreshTokensParams{
		PublicID: userID,
		Limit:   20,
		Offset:  0,
	}) // TODO: ページネーション
	if err != nil {
		return nil, fmt.Errorf("GetRefreshTokens: %w", err)
	}

	domainSessions := make([]Session, len(sessions))
	for i, session := range sessions {
		domainSessions[i] = NewSession(
			session.PublicID,
			session.CreatedAt,
			session.UpdatedAt,
			session.CountryCode,
			session.CityName,
			session.BrowserName,
		)
	}

	return domainSessions, nil
}

func (s *Service) RemoveSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	userId, err := util.MustUserIDFromContext(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	q := sqlc.New(s.db)
	removedPublicId, err := q.RemoveRefreshTokenByIDAndSaveJtiBlacklist(ctx, sqlc.RemoveRefreshTokenByIDAndSaveJtiBlacklistParams{
		RefreshTokenPublicID: sessionID,
		ExpiresAt:            time.Now().UTC().Add(s.jwtService.TokenDuration()).UTC(),
		UserPublicID:         userId,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoSuchRefreshToken
		}
		return uuid.Nil, fmt.Errorf("RemoveRefreshTokenByIDAndSaveJtiBlacklist: %w", err)
	}

	return removedPublicId, nil
}
