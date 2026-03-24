package auth

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
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

func (a *authorizationRepositoryImpl) Save(ctx context.Context, authorization *Authorization) (_ int64, err error) {
	defer util.Wrap(&err, "authorizationRepository.Save")
	saveAuthorization, err := a.q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         authorization.Issuer,
		Sub:            authorization.Sub,
		LastLoggedInAt: authorization.LastLoggedInAt,
		PublicID:       authorization.ID,
	})
	if err != nil {
		return 0, err
	}

	return saveAuthorization.MUserAuthorizationID, nil
}

var _ AuthorizationRepository = (*authorizationRepositoryImpl)(nil)

type RefreshTokenRepository interface {
	Save(ctx context.Context, authorizationID int64, refreshToken *RefreshToken) (int64, error)
	RevokeByTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, jtiExpiresAt time.Time) error
	RevokeByID(ctx context.Context, userID, sessionID uuid.UUID, jtiExpiresAt time.Time) (_ uuid.UUID, err error)
	RotateRefreshToken(ctx context.Context, newRefreshToken *RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (_ uuid.UUID, err error)
}

type refreshTokenRepositoryImpl struct {
	q sqlc.Querier
}

func NewRefreshTokenRepository(q sqlc.Querier) RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{
		q: q,
	}
}

func (r *refreshTokenRepositoryImpl) Save(ctx context.Context, authorizationID int64, refreshToken *RefreshToken) (_ int64, err error) {
	defer util.Wrap(&err, "refreshTokenRepository.Save(authorizationID=%d)", authorizationID)
	refreshTokenID, err := r.q.InsertRefreshToken(ctx, sqlc.InsertRefreshTokenParams{
		MUserAuthorizationID: authorizationID,
		TokenHash:            refreshToken.TokenHash,
		Generation:           1,
		PublicID:             refreshToken.ID,
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
		return 0, err
	}

	return refreshTokenID, nil
}

func (r *refreshTokenRepositoryImpl) RevokeByTokenHash(ctx context.Context, userID uuid.UUID, tokenHash string, jtiExpiresAt time.Time) (err error) {
	defer util.Wrap(&err, "refreshTokenRepository.RevokeByTokenHash(userID=%s)", userID)
	if _, err := r.q.RevokeRefreshTokenByHash(ctx, sqlc.RevokeRefreshTokenByHashParams{
		UserPublicID: userID,
		TokenHash:    tokenHash,
		ExpiresAt:    jtiExpiresAt,
	}); err != nil {
		return err
	}
	return nil
}

func (r *refreshTokenRepositoryImpl) RevokeByID(ctx context.Context, userID, sessionID uuid.UUID, jtiExpiresAt time.Time) (_ uuid.UUID, err error) {
	defer util.Wrap(&err, "refreshTokenRepository.RevokeByID(userID=%s, sessionID=%s)", userID, sessionID)
	removedPublicID, err := r.q.RevokeRefreshTokenByID(ctx, sqlc.RevokeRefreshTokenByIDParams{
		RefreshTokenPublicID: sessionID,
		ExpiresAt:            jtiExpiresAt,
		UserPublicID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, err
		}
		return uuid.Nil, err
	}
	return removedPublicID, nil
}

func (r *refreshTokenRepositoryImpl) RotateRefreshToken(ctx context.Context, newRefreshToken *RefreshToken, tokenHashForCheck string, updatedAtForCheck time.Time) (_ uuid.UUID, err error) {
	defer util.Wrap(&err, "refreshTokenRepository.RotateRefreshToken")
	var userID uuid.UUID
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
		return uuid.Nil, err
	}
	return userID, nil
}

var _ RefreshTokenRepository = (*refreshTokenRepositoryImpl)(nil)
