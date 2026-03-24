package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthorizationQueryService interface{}

type authorizationQueryServiceImpl struct {
	q sqlc.Querier
}

func NewAuthorizationQueryService(db *pgxpool.Pool) AuthorizationQueryService {
	return &authorizationQueryServiceImpl{
		q: sqlc.New(db),
	}
}

var _ AuthorizationQueryService = (*authorizationQueryServiceImpl)(nil)

type GetSessionsView struct {
	ID             uuid.UUID
	ActivatedAt    time.Time
	LastLoggedInAt time.Time
	UserAgent      string
	IpAddress      string
	CountryCode    string
	CityName       string
	BrowserName    string
	DeviceType     string
}

type RefreshTokenQueryService interface {
	GetSessions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSessionsView, error)
}

type refreshTokenQueryServiceImpl struct {
	q sqlc.Querier
}

func NewRefreshTokenQueryService(db *pgxpool.Pool) RefreshTokenQueryService {
	return &refreshTokenQueryServiceImpl{
		q: sqlc.New(db),
	}
}

func (r *refreshTokenQueryServiceImpl) GetSessions(ctx context.Context, userID uuid.UUID, cursor *uuid.UUID, limit int32) ([]GetSessionsView, error) {
	rows, err := r.q.ListRefreshTokens(ctx, sqlc.ListRefreshTokensParams{
		UserID:     userID,
		QueryLimit: limit,
		Cursor:     cursor,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh tokens(refreshTokenQueryService.GetSessions): %w", err)
	}

	views := make([]GetSessionsView, len(rows))
	for i, row := range rows {
		views[i] = GetSessionsView{
			ID:             row.PublicID,
			ActivatedAt:    row.ActivatedAt,
			LastLoggedInAt: row.LastLoggedInAt,
			UserAgent:      row.UserAgent,
			IpAddress:      row.IpAddress,
			CountryCode:    row.CountryCode,
			CityName:       row.CityName,
			BrowserName:    row.BrowserName,
			DeviceType:     row.DeviceType,
		}
	}

	return views, nil
}

var _ RefreshTokenQueryService = (*refreshTokenQueryServiceImpl)(nil)
