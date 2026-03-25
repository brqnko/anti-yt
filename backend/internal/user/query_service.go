package user

import (
	"context"
	"errors"
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
	StartTime string
	EndTime   string
}

type UserQueryService interface {
	Find(ctx context.Context, userID uuid.UUID) (UserStatusView, error)
	FindByAuthorizationID(ctx context.Context, authorizationID int64) (_ uuid.UUID, _ bool, err error)
}

type userQueryServiceImpl struct {
	q sqlc.Querier
}

func NewUserQueryService(db *pgxpool.Pool) UserQueryService {
	return &userQueryServiceImpl{q: sqlc.New(db)}
}

func (u *userQueryServiceImpl) Find(ctx context.Context, userID uuid.UUID) (_ UserStatusView, err error) {
	defer util.Wrap(&err, "userQueryService.Find(userID=%s)", userID)

	rows, err := u.q.GetUserProfile(ctx, userID)
	if err != nil {
		return UserStatusView{}, err
	}
	if len(rows) == 0 {
		return UserStatusView{}, pgx.ErrNoRows
	}

	first := rows[0]

	// rangeが0件の場合はNULL行1行のみ返る
	var ranges []ScreenTimeLimitRangeView
	if first.ScreenTimeRangeStart != nil {
		ranges = make([]ScreenTimeLimitRangeView, len(rows))
		for i, row := range rows {
			startTime, err := util.IntToHHmm(int(*row.ScreenTimeRangeStart) / 60)
			if err != nil {
				return UserStatusView{}, err
			}
			endTime, err := util.IntToHHmm(int(*row.ScreenTimeRangeEnd) / 60)
			if err != nil {
				return UserStatusView{}, err
			}
			ranges[i] = ScreenTimeLimitRangeView{
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

func (u *userQueryServiceImpl) FindByAuthorizationID(ctx context.Context, authorizationID int64) (_ uuid.UUID, _ bool, err error) {
	defer util.Wrap(&err, "userQueryService.FindByAuthorizationID(authorizationID=%d)", authorizationID)

	row, err := u.q.GetUserIDByAuthorization(ctx, authorizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, err
		}
		return uuid.Nil, false, err
	}
	return row.PublicID, row.IsDeactivated, nil
}

var _ UserQueryService = (*userQueryServiceImpl)(nil)
