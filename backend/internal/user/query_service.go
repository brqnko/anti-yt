package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStatusView struct {
	UserID               uuid.UUID
	DisplayName          string
	LanguageCode         string
	JoinedAt             time.Time
	DailyScreenSeconds   *int
	ScreenTimeLimitRange []ScreenTimeLimitRangeView
}

type ScreenTimeLimitRangeView struct {
	ID        uuid.UUID
	StartTime string
	EndTime   string
}

type UserQueryService interface {
	Find(ctx context.Context, userID uuid.UUID) (UserStatusView, error)
	FindByAuthorizationID(ctx context.Context, authorizationID int64) (userID uuid.UUID, isDeactivated bool, err error)
}

type userQueryServiceImpl struct {
	q sqlc.Querier
}

func NewUserQueryService(db *pgxpool.Pool) UserQueryService {
	return &userQueryServiceImpl{q: sqlc.New(db)}
}

func (u *userQueryServiceImpl) Find(ctx context.Context, userID uuid.UUID) (UserStatusView, error) {
	rows, err := u.q.GetUserProfile(ctx, userID)
	if err != nil {
		return UserStatusView{}, fmt.Errorf("failed to getUserWithScreenTimeRanges(userQueryService.Find): %w", err)
	}
	if len(rows) == 0 {
		return UserStatusView{}, pgx.ErrNoRows
	}

	first := rows[0]

	// rangeが0件の場合はNULL行1行のみ返る
	var ranges []ScreenTimeLimitRangeView
	if first.ScreenTimeRangeID != nil {
		ranges = make([]ScreenTimeLimitRangeView, len(rows))
		for i, row := range rows {
			startTime, err := util.IntToHHmm(int(*row.ScreenTimeRangeStart))
			if err != nil {
				return UserStatusView{}, fmt.Errorf("failed to convert startTime(userQueryService.Find): %w", err)
			}
			endTime, err := util.IntToHHmm(int(*row.ScreenTimeRangeEnd))
			if err != nil {
				return UserStatusView{}, fmt.Errorf("failed to convert endTime(userQueryService.Find): %w", err)
			}
			ranges[i] = ScreenTimeLimitRangeView{
				ID:        *row.ScreenTimeRangeID,
				StartTime: startTime,
				EndTime:   endTime,
			}
		}
	}

	var dailyScreenSeconds *int
	if !IsUnlimitedScreenTimeSeconds(first.DailyScreenTimeSeconds) {
		dailyScreenSeconds = &first.DailyScreenTimeSeconds
	}

	return UserStatusView{
		UserID:               userID,
		DisplayName:          first.DisplayName,
		LanguageCode:         first.LanguageCode,
		JoinedAt:             first.JoinedAt.UTC(),
		DailyScreenSeconds:   dailyScreenSeconds,
		ScreenTimeLimitRange: ranges,
	}, nil
}

func (u *userQueryServiceImpl) FindByAuthorizationID(ctx context.Context, authorizationID int64) (uuid.UUID, bool, error) {
	row, err := u.q.GetUserIDByAuthorization(ctx, authorizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, err
		}
		return uuid.Nil, false, fmt.Errorf("failed to userQueryService.FindByAuthorizationID: %w", err)
	}
	return row.PublicID, row.IsDeactivated, nil
}

var _ UserQueryService = (*userQueryServiceImpl)(nil)
