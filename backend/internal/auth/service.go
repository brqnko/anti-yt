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
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mssola/user_agent"
	"golang.org/x/oauth2"
)

var (
	ErrCSRFOrStateIsEmpty = errors.New("csrf or state is empty")
	ErrCSRFIsWrong        = errors.New("csrf != state")
	ErrIDTokenNotFound    = errors.New("id token not found")

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
	AccessToken  string
	RefreshToken string
	CSRFToken    string
	IsCreated    bool
}

type Service struct {
	db *pgxpool.Pool

	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier

	jwtService jwt_d.JWTService

	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration

	serverURL string
}

func NewService(db *pgxpool.Pool, oauth2Config *oauth2.Config, verifier *oidc.IDTokenVerifier, accessTokenDuration, refreshTokenDuration time.Duration, serverURL string, jwtService jwt_d.JWTService) (*Service, error) {
	return &Service{
		db:                   db,
		oauth2Config:         oauth2Config,
		verifier:             verifier,
		jwtService:           jwtService,
		AccessTokenDuration:  accessTokenDuration,
		RefreshTokenDuration: refreshTokenDuration,
		serverURL:            serverURL,
	}, nil
}

func (s *Service) CreateAuthCode(ctx context.Context) (redirectURL, csrf string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	csrf = base64.URLEncoding.EncodeToString(b)

	return s.oauth2Config.AuthCodeURL(csrf), csrf, nil
}

func (s *Service) GoogleOIDCCallback(ctx context.Context, params GoogleOIDCCallbackParams) (*GoogleOIDCCallbackResult, error) {
	if params.CSRF == "" || params.State == "" {
		return nil, ErrCSRFOrStateIsEmpty
	}
	if params.CSRF != params.State {
		return nil, ErrCSRFIsWrong
	}

	oauth2Token, err := s.oauth2Config.Exchange(ctx, params.Code)
	if err != nil {
		return nil, err
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, ErrIDTokenNotFound
	}
	idToken, err := s.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, err
	}
	var oidcClaims struct {
		Sub string `json:"sub"`
	}
	if err := idToken.Claims(&oidcClaims); err != nil {
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
		Sub:    oidcClaims.Sub,
	})
	if err != nil {
		return nil, err
	}

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	tokenString := base64.RawURLEncoding.EncodeToString(token)
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashString := hex.EncodeToString(tokenHash[:])

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
		ExpiresAt:            time.Now().UTC().Add(s.RefreshTokenDuration),
	}); err != nil {
		return nil, err
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	var accessTokenString string
	if !authorization.IsCreated {
		userId, err := q.GetUserIDByAuthorization(ctx, authorization.MUserAuthorizationID)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
		} else {
			// すでに認証テーブルに情報があり、かつ、ユーザーテーブルに存在する
			accessTokenString, err = s.jwtService.SignUserAccessToken(userId, jti, s.serverURL, s.AccessTokenDuration)
			if err != nil {
				return nil, err
			}
		}
	}
	if accessTokenString == "" {
		accessTokenString, err = s.jwtService.SignRegisterToken(authorization.PublicID, jti, s.serverURL, s.AccessTokenDuration)
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
		AccessToken:  accessTokenString,
		RefreshToken: tokenString,
		CSRFToken:    csrfToken,
		IsCreated:    authorization.IsCreated,
	}, nil
}

func (s *Service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	_, jti, expiresAt, err := s.jwtService.VerifyUserAccessTokenWithExpiry(accessToken)
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

func (s *Service) RefreshToken(ctx context.Context, refreshToken, ipAddress, countryCode, deviceFingerprint, userAgent string) (newRefreshToken, accessToken string, err error) {
	refreshTokenHash := sha256.Sum256([]byte(refreshToken))
	refreshTokenHashString := hex.EncodeToString(refreshTokenHash[:])

	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return "", "", err
	}
	tokenString := base64.RawURLEncoding.EncodeToString(token)
	tokenHash := sha256.Sum256([]byte(tokenString))
	tokenHashString := hex.EncodeToString(tokenHash[:])

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", "", err
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
		ExpiresAt:         time.Now().UTC().Add(s.RefreshTokenDuration),
		IpAddress:         ipAddress,
		DeviceFingerprint: deviceFingerprint,
		UserAgent:         userAgent,
		CountryCode:       countryCode,
		CityName:          "", // TODO
		BrowserName:       fmt.Sprintf("%s%s", browserName, browserVersion),
		DeviceType:        ua.OSInfo().FullName,
		TokenHash_2:       refreshTokenHashString,                       // 昔のRefreshToken
		ExpiresAt_2:       time.Now().UTC(),                             // RefreshTokenの有効期限
		UpdatedAt:         time.Now().UTC().Add(-s.AccessTokenDuration), // RefreshTokenをたくさん発行されるのを防ぐため、updatedAt + accessTokenDuration < now => updatedAt < now - accessTokenDuration
	})
	if err != nil {
		return "", "", err
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return "", "", err
	}
	userID, err := q.GetUserIDByAuthorization(ctx, authorizationID)
	if err != nil {
		return "", "", err
	}
	accessTokenString, err := s.jwtService.SignUserAccessToken(userID, jti, s.serverURL, s.AccessTokenDuration)
	if err != nil {
		return "", "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", "", err
	}

	return tokenString, accessTokenString, nil
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
