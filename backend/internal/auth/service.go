package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidCSRFOrState = errors.New("invalid csrf or state")
	ErrInvalidCSRF        = errors.New("invalid csrf: csrf != state")
	ErrIDTokenNotFound    = oidc.ErrIDTokenNotFound

)

type Service struct {
	db *pgxpool.Pool

	oidcService oidc.GoogleOIDCService

	jwtService jwt_d.JWTService

	serverURL            string
	refreshTokenDuration time.Duration

	refreshTokenQS  RefreshTokenQueryService
	authorizationQS AuthorizationQueryService
	userQS          user.UserQueryService
}

func NewService(db *pgxpool.Pool, oidcService oidc.GoogleOIDCService, serverURL string, jwtService jwt_d.JWTService, refreshTokenDuration time.Duration) (*Service, error) {
	return &Service{
		db:                        db,
		oidcService:               oidcService,
		jwtService:                jwtService,
		serverURL:                 serverURL,
		refreshTokenDuration:      refreshTokenDuration,
		refreshTokenQS:  NewRefreshTokenQueryService(db),
		authorizationQS: NewAuthorizationQueryService(db),
		userQS:          user.NewUserQueryService(db),
	}, nil
}

func (s *Service) CreateAuthCode(ctx context.Context) (redirectURL, csrf string, err error) {
	csrfToken, err := util.RandomStringUrlSafe(32)
	if err != nil {
		return "", "", err
	}

	return s.oidcService.AuthCodeURL(csrfToken), csrfToken, nil
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, csrf, state, code, ipAddress, countryCode, deviceFingerprint, userAgent string) (accessToken, refreshTokenRaw, csrfToken, redirectPath string, accessTokenExpiresAt, refreshTokenExpiresAt time.Time, err error) {
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
			slog.Error("failed to rollback", "error", err)
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

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	q := sqlc.New(s.db)
	userID, _, _, _ := s.jwtService.VerifyUserAccessToken(accessToken)

	refreshTokenHash := util.Sha256Hex(refreshToken)
	return NewRefreshTokenRepository(q).RevokeByTokenHash(ctx, userID, refreshTokenHash, time.Now().UTC().Add(s.refreshTokenDuration))
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, deviceFingerprint, userAgent string) (refreshTokenRaw string, accessToken string, accessTokenExpiresAt, refreshTokenExpiresAt time.Time, err error) {
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
			slog.Error("failed to rollback", "error", err)
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

func (s *Service) GetSessions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) (sessions []GetSessionsView, hasNext bool, err error) {
	view, err := s.refreshTokenQS.GetSessions(ctx, userID, cursor, limit+1)
	if err != nil {
		return nil, false, err
	}
	if len(view) > int(limit) {
		return view[:limit], true, nil
	}
	return view, false, nil
}

func (s *Service) RemoveSession(ctx context.Context, userID, sessionID uuid.UUID) (uuid.UUID, error) {
	q := sqlc.New(s.db)
	removedPublicID, err := NewRefreshTokenRepository(q).RevokeByID(ctx, userID, sessionID, time.Now().UTC().Add(s.jwtService.TokenDuration()))
	if err != nil {
		return uuid.Nil, err
	}

	return removedPublicID, nil
}
