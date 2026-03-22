package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
)

type AuthorizationRepository interface {
	Save(ctx context.Context, authorization Authorization) (int64, error)
	FindUserByAuthorizationID(ctx context.Context, authorizationID int64) (userID uuid.UUID, isDeactivated bool, err error)
}

func NewAuthorizationRepository(q sqlc.Querier) AuthorizationRepository {
	return &authorizationRepositoryImpl{
		q: q,
	}
}

type authorizationRepositoryImpl struct {
	q sqlc.Querier
}

func (a *authorizationRepositoryImpl) Save(ctx context.Context, authorization Authorization) (int64, error) {
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

func (a *authorizationRepositoryImpl) FindUserByAuthorizationID(ctx context.Context, authorizationID int64) (userID uuid.UUID, isDeactivated bool, err error) {
	row, err := a.q.GetUserIDByAuthorization(ctx, authorizationID)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("failed to authorizationRepository.FindUserByAuthorizationID: %w", err)
	}
	return row.PublicID, row.IsDeactivated, nil
}

var _ AuthorizationRepository = (*authorizationRepositoryImpl)(nil)

type RefreshTokenRepository interface {
	Save(ctx context.Context, authorizationID int64, refreshToken RefreshToken) (int64, error)
	RevokeByTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, jtiExpiresAt time.Time) error
	RevokeByID(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, jtiExpiresAt time.Time) (uuid.UUID, error)
	RotateRefreshToken(ctx context.Context, newRefreshToken RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (userID uuid.UUID, err error)
	GetRefreshTokens(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]RefreshToken, error)
}

type refreshTokenRepositoryImpl struct {
	q sqlc.Querier
}

func NewRefreshTokenRepository(q sqlc.Querier) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{
		q: q,
	}
}

func (r *refreshTokenRepositoryImpl) Save(ctx context.Context, authorizationID int64, refreshToken RefreshToken) (int64, error) {
	refreshTokenID, err := r.q.SaveRefreshToken(ctx, sqlc.SaveRefreshTokenParams{
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
	if _, err := r.q.RemoveRefreshTokenByTokenHashAndSaveJtiBlacklist(ctx, sqlc.RemoveRefreshTokenByTokenHashAndSaveJtiBlacklistParams{
		UserPublicID: userID,
		TokenHash:    tokenHash,
		ExpiresAt:    jtiExpiresAt,
	}); err != nil {
		return fmt.Errorf("failed to refreshTokenRepository.RevokeByTokenHash: %w", err)
	}
	return nil
}

func (r *refreshTokenRepositoryImpl) RevokeByID(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, jtiExpiresAt time.Time) (uuid.UUID, error) {
	removedPublicID, err := r.q.RemoveRefreshTokenByIDAndSaveJtiBlacklist(ctx, sqlc.RemoveRefreshTokenByIDAndSaveJtiBlacklistParams{
		RefreshTokenPublicID: sessionID,
		ExpiresAt:            jtiExpiresAt,
		UserPublicID:         userID,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to refreshTokenRepository.RevokeByID: %w", err)
	}
	return removedPublicID, nil
}

func (r *refreshTokenRepositoryImpl) RotateRefreshToken(ctx context.Context, newRefreshToken RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (userID uuid.UUID, err error) {
	userID, err = r.q.UpdateRefreshToken(ctx, sqlc.UpdateRefreshTokenParams{
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

func (r *refreshTokenRepositoryImpl) GetRefreshTokens(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]RefreshToken, error) {
	tokens, err := r.q.GetRefreshTokens(ctx, sqlc.GetRefreshTokensParams{
		PublicID: userID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to refreshTokenRepository.GetRefreshTokens: %w", err)
	}

	refreshTokens := make([]RefreshToken, len(tokens))
	for i, token := range tokens {
		refreshToken, err := NewRefreshToken(
			token.UserAgent,
			token.DeviceFingerprint,
			token.IpAddress,
			token.CountryCode,
			token.CityName,
			token.ExpiresAt,
			WithRefreshTokenActivatedAt(token.ActivatedAt),
			WithRefreshTokenHash(token.TokenHash),
			WithRefreshTokenID(token.PublicID),
			WithRefreshTokenLastLoggedInAt(token.LastLoggedInAt),
		)
		if err != nil {
			return nil, err
		}
		refreshTokens[i] = refreshToken
	}

	return refreshTokens, nil
}

var _ RefreshTokenRepository = (*refreshTokenRepositoryImpl)(nil)
