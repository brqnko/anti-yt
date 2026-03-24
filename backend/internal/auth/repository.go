package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthorizationRepository interface {
	Save(ctx context.Context, authorization *Authorization) (int64, error)
}

func NewAuthorizationRepository(q sqlc.Querier) AuthorizationRepository {
	return &authorizationRepositoryImpl{
		q: q,
	}
}

type authorizationRepositoryImpl struct {
	q sqlc.Querier
}

func (a *authorizationRepositoryImpl) Save(ctx context.Context, authorization *Authorization) (int64, error) {
	saveAuthorization, err := a.q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         authorization.Issuer,
		Sub:            authorization.Sub,
		LastLoggedInAt: authorization.LastLoggedInAt,
		PublicID:       authorization.ID,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to authorizationRepository.Save: %w", err)
	}

	return saveAuthorization.MUserAuthorizationID, nil
}

var _ AuthorizationRepository = (*authorizationRepositoryImpl)(nil)

type RefreshTokenRepository interface {
	Save(ctx context.Context, authorizationID int64, refreshToken *RefreshToken) (int64, error)
	RevokeByTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, jtiExpiresAt time.Time) error
	RevokeByID(ctx context.Context, userID, sessionID uuid.UUID, jtiExpiresAt time.Time) (id uuid.UUID, err error)
	RotateRefreshToken(ctx context.Context, newRefreshToken *RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (userID uuid.UUID, err error)
}

type refreshTokenRepositoryImpl struct {
	q sqlc.Querier
}

func NewRefreshTokenRepository(q sqlc.Querier) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{
		q: q,
	}
}

func (r *refreshTokenRepositoryImpl) Save(ctx context.Context, authorizationID int64, refreshToken *RefreshToken) (int64, error) {
	refreshTokenID, err := r.q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
		MUserAuthorizationID: authorizationID,
		TokenHash:            refreshToken.TokenHash,
		IpAddress:            refreshToken.IpAddress,
		DeviceFingerprint:    refreshToken.DeviceFingerprint,
		UserAgent:            refreshToken.UserAgent,
		CountryCode:          refreshToken.CountryCode,
		CityName:             refreshToken.CityName,
		BrowserName:          refreshToken.BrowserName,
		DeviceType:           refreshToken.DeviceType,
		ExpiresAt:            refreshToken.ExpiresAt,
		AccessTokenJti:       refreshToken.AccessTokenJTI,
		ActivatedAt:          refreshToken.ActivatedAt,
		LastLoggedInAt:       refreshToken.LastLoggedInAt,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to refreshTokenRepository.Save: %w", err)
	}

	return refreshTokenID, nil
}

func (r *refreshTokenRepositoryImpl) RevokeByTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, jtiExpiresAt time.Time) error {
	if _, err := r.q.RevokeRefreshTokenByHash(ctx, sqlc.RevokeRefreshTokenByHashParams{
		UserPublicID: userID,
		TokenHash:    tokenHash,
		ExpiresAt:    jtiExpiresAt,
	}); err != nil {
		return fmt.Errorf("failed to refreshTokenRepository.RevokeByTokenHash: %w", err)
	}
	return nil
}

func (r *refreshTokenRepositoryImpl) RevokeByID(ctx context.Context, userID, sessionID uuid.UUID, jtiExpiresAt time.Time) (uuid.UUID, error) {
	removedPublicID, err := r.q.RevokeRefreshTokenByID(ctx, sqlc.RevokeRefreshTokenByIDParams{
		RefreshTokenPublicID: sessionID,
		ExpiresAt:            jtiExpiresAt,
		UserPublicID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, err
		}
		return uuid.Nil, fmt.Errorf("failed to refreshTokenRepository.RevokeByID: %w", err)
	}
	return removedPublicID, nil
}

func (r *refreshTokenRepositoryImpl) RotateRefreshToken(ctx context.Context, newRefreshToken *RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (userID uuid.UUID, err error) {
	userID, err = r.q.RotateRefreshToken(ctx, sqlc.RotateRefreshTokenParams{
		NewTokenHash:         newRefreshToken.TokenHash,
		NewExpiresAt:         newRefreshToken.ExpiresAt,
		NewIpAddress:         newRefreshToken.IpAddress,
		NewDeviceFingerprint: newRefreshToken.DeviceFingerprint,
		NewUserAgent:         newRefreshToken.UserAgent,
		NewCountryCode:       newRefreshToken.CountryCode,
		NewCityName:          newRefreshToken.CityName,
		NewBrowserName:       newRefreshToken.BrowserName,
		NewDeviceType:        newRefreshToken.DeviceType,
		NewAccessTokenJti:    newRefreshToken.AccessTokenJTI,
		TokenHashForCheck:    tokenHashForCheck,
		UpdatedAtForCheck:    updatedAtForCheck,
		LastLoggedInAt:       newRefreshToken.LastLoggedInAt,
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to refreshTokenRepository.RotateRefreshToken: %w", err)
	}
	return userID, nil
}

var _ RefreshTokenRepository = (*refreshTokenRepositoryImpl)(nil)
