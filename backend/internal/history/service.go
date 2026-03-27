package history

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool

	historyQS HistoryQueryService
}

func NewService(db *pgxpool.Pool) *Service {

	return &Service{
		db:        db,
		historyQS: NewHistoryQueryService(db),
	}
}

func (s *Service) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int, loc *time.Location) (_ *int, err error) {
	defer util.Wrap(&err, "Service.Heartbeat")

	if err := NewHistoryRepository(sqlc.New(s.db)).Heartbeat(ctx, userID, videoID, positionSeconds); err != nil {
		return nil, err
	}

	watchStats, err := s.historyQS.FindTotalWatchSeconds(ctx, userID, loc)
	if err != nil {
		return nil, err
	}

	if user.IsUnlimitedScreenTimeSeconds(watchStats.DailyLimitSeconds) {
		return nil, nil
	}
	remaining := max(0, watchStats.DailyLimitSeconds-watchStats.TodayWatchTotal)
	return &remaining, nil
}

func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, limit int, cursor *uuid.UUID, loc *time.Location) (_ []GetHistoryView, _ bool, err error) {
	defer util.Wrap(&err, "Service.GetHistory")

	views, err := s.historyQS.FindHistory(ctx, userID, cursor, int32(limit+1))
	if err != nil {
		return nil, false, err
	}

	for i := range views {
		views[i].WatchedAt = views[i].WatchedAt.In(loc)
	}

	if len(views) > limit {
		return views[:limit], true, nil
	}
	return views, false, nil
}

func (s *Service) GetStatisticsByWeek(ctx context.Context, userID uuid.UUID, targetWeek time.Time, loc *time.Location) (_ *string, _ []GetStatisticsWeeklyView, err error) {
	defer util.Wrap(&err, "Service.GetStatisticsByWeek")

	aiSummary, views, err := s.historyQS.FindStatisticsByWeek(ctx, userID, targetWeek)
	if err != nil {
		return nil, nil, err
	}

	for i := range views {
		views[i].WatchDate = views[i].WatchDate.In(loc)
	}

	return aiSummary, views, nil
}
