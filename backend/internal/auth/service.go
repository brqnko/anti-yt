package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/core/oidc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mssola/user_agent"
)

var (
	ErrCSRFOrStateIsEmpty = errors.New("csrf or state is empty")
	ErrCSRFIsWrong        = errors.New("csrf != state")
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
	IsCreated             bool
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
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	csrf = base64.URLEncoding.EncodeToString(b)

	return s.oidcService.AuthCodeURL(csrf), csrf, nil
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, params GoogleOIDCCallbackParams) (*GoogleOIDCCallbackResult, error) {
	if params.CSRF == "" || params.State == "" {
		return nil, ErrCSRFOrStateIsEmpty
	}
	if params.CSRF != params.State {
		return nil, ErrCSRFIsWrong
	}

	sub, err := s.oidcService.ExchangeAndVerify(ctx, params.Code)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	authorization, err := q.CreateAuthorization(ctx, sqlc.CreateAuthorizationParams{
		Issuer: "https://accounts.google.com",
		Sub:    sub,
	})
	if err != nil {
		return nil, err
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	tokenString := base64.URLEncoding.EncodeToString(token)
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashString := hex.EncodeToString(tokenHash[:])

	refreshTokenExpiresAt := time.Now().UTC().Add(s.refreshTokenDuration)

	ua := user_agent.New(params.UserAgent)
	browserName, browserVersion := ua.Browser()
	if err := q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		MUserAuthorizationID: authorization.MUserAuthorizationID,
		TokenHash:            tokenHashString,
		IpAddress:            params.IPAddress,
		DeviceFingerprint:    params.DeviceFingerprint,
		UserAgent:            params.UserAgent,
		CountryCode:          params.CountryCode,
		CityName:             "", // TODO: 今はなし
		BrowserName:          fmt.Sprintf("%s%s", browserName, browserVersion),
		DeviceType:           ua.OSInfo().FullName,
		ExpiresAt:            refreshTokenExpiresAt.UTC(),
	}); err != nil {
		return nil, err
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	var accessTokenString string
	var accessTokenExpiresAt time.Time
	if !authorization.IsCreated {
		userId, err := q.GetUserIDByAuthorization(ctx, authorization.MUserAuthorizationID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			// すでに認証テーブルに情報があり、かつ、ユーザーテーブルに存在する
			accessTokenString, accessTokenExpiresAt, err = s.jwtService.SignUserAccessToken(userId, jti, s.serverURL)
			if err != nil {
				return nil, err
			}
		}
	}
	if accessTokenString == "" {
		accessTokenString, accessTokenExpiresAt, err = s.jwtService.SignRegisterToken(authorization.PublicID, jti, s.serverURL)
		if err != nil {
			return nil, err
		}
	}

	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	csrfToken := base64.URLEncoding.EncodeToString(b)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &GoogleOIDCCallbackResult{
		AccessToken:           accessTokenString,
		RefreshToken:          tokenString,
		CSRFToken:             csrfToken,
		IsCreated:             authorization.IsCreated,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}, nil
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	_, jti, expiresAt, err := s.jwtService.VerifyUserAccessToken(accessToken)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	refreshTokenHash := sha256.Sum256([]byte(refreshToken))
	refreshTokenString := hex.EncodeToString(refreshTokenHash[:])

	if _, err := q.RemoveRefreshToken(ctx, refreshTokenString); err != nil {
		return err
	}

	if err := q.SaveJTIBlacklist(ctx, sqlc.SaveJTIBlacklistParams{
		Jti:       jti,
		ExpiresAt: expiresAt,
	}); err != nil {
		// NOTE: ユーザーはlogoutを望んでいるのでブラックリストの更新の失敗で、でrollbackすべきではない
		slog.Error("failed to save jti into blacklist", "error", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, deviceFingerprint, userAgent string) (newRefreshToken, accessToken string, accessTokenExpiresAt, refreshTokenExpiresAt time.Time, err error) {
	refreshTokenHash := sha256.Sum256([]byte(refreshToken))
	refreshTokenHashString := hex.EncodeToString(refreshTokenHash[:])

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	tokenString := base64.URLEncoding.EncodeToString(token)
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashString := hex.EncodeToString(tokenHash[:])

	now := time.Now().UTC()
	rtExp := now.Add(s.refreshTokenDuration)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, sql.ErrTxDone) {
			slog.Error("failed to rollback", "error", err)
		}
	}()
	q := sqlc.New(tx)

	ua := user_agent.New(userAgent)
	browserName, browserVersion := ua.Browser()
	authorizationID, err := q.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
		TokenHash:         tokenHashString,
		ExpiresAt:         rtExp.UTC(),
		IpAddress:         ipAddress,
		DeviceFingerprint: deviceFingerprint,
		UserAgent:         userAgent,
		CountryCode:       countryCode,
		CityName:          "", // TODO
		BrowserName:       fmt.Sprintf("%s%s", browserName, browserVersion),
		DeviceType:        ua.OSInfo().FullName,
		TokenHash_2:       refreshTokenHashString,                 // 昔のRefreshToken
		ExpiresAt_2:       now,                                    // RefreshTokenの有効期限
		UpdatedAt:         now.Add(-s.jwtService.TokenDuration()), // RefreshTokenをたくさん発行されるのを防ぐため、updatedAt + accessTokenDuration < now => updatedAt < now - accessTokenDuration
	})
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	userID, err := q.GetUserIDByAuthorization(ctx, authorizationID)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}
	accessTokenString, signedAtExp, err := s.jwtService.SignUserAccessToken(userID, jti, s.serverURL)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return tokenString, accessTokenString, signedAtExp, rtExp, nil
}

func (s *Service) GetSessions(ctx context.Context, userID uuid.UUID) ([]Session, error) {
	q := sqlc.New(s.db)

	mUserID, err := q.GetUserAuthorizationID(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessions, err := q.GetUserAllRefreshTokens(ctx, mUserID)
	if err != nil {
		return nil, err
	}

	domainSessions := make([]Session, len(sessions))
	for i, session := range sessions {
		domainSessions[i] = Session{
			ID:             session.PublicID,
			CreatedAt:      session.CreatedAt,
			LastLoggedInAt: session.UpdatedAt,
			CountryCode:    session.CountryCode,
			CityName:       session.CityName,
			BrowserName:    session.BrowserName,
		}
	}

	return domainSessions, nil
}

func (s *Service) RemoveSession(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, error) {
	q := sqlc.New(s.db)
	tokenID, err := q.RemoveRefreshTokenByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNoSuchRefreshToken
		}
		return uuid.Nil, err
	}

	// TODO: jti と expiresAt が未定義のため一時的にコメントアウト
	// if err := q.SaveJTIBlacklist(ctx, sqlc.SaveJTIBlacklistParams{
	// 	Jti:       jti,
	// 	ExpiresAt: expiresAt,
	// }); err != nil {
	// 	// NOTE: ユーザーはlogoutを望んでいるのでブラックリストの更新の失敗で、でrollbackすべきではない
	// 	slog.Error("failed to save jti into blacklist", "error", err)
	// }

	return tokenID, nil
}
