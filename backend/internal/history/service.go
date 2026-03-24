package history

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db        *pgxpool.Pool
	historyQS HistoryQueryService
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{
		db:        db,
		historyQS: NewHistoryQueryService(db),
	}
}

func (s *Service) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int) (*int, error) {
	if err := NewHistoryRepository(sqlc.New(s.db)).Heartbeat(ctx, userID, videoID, positionSeconds); err != nil {
		return nil, err
	}

	watchStats, err := s.historyQS.FindTotalWatchSeconds(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.IsUnlimitedScreenTimeSeconds(watchStats.DailyLimitSeconds) {
		return nil, nil
	}
	remaining := max(0, watchStats.DailyLimitSeconds-watchStats.TodayWatchTotal)
	return &remaining, nil
}

func (s *Service) GetHistory(ctx context.Context, userID uuid.UUID, limit int, cursor *uuid.UUID) (_ []GetHistoryView, _ bool, err error) {
	views, err := s.historyQS.FindHistory(ctx, userID, cursor, int32(limit+1))
	if err != nil {
		return nil, false, err
	}

	if len(views) > limit {
		return views[:limit], true, nil
	}
	return views, false, nil
}

func (s *Service) GetStatisticsByWeek(ctx context.Context, userID uuid.UUID, targetWeek time.Time) ([]GetStatisticsWeeklyView, error) {
	views, err := s.historyQS.FindStatisticsByWeek(ctx, userID, targetWeek)
	if err != nil {
		return nil, err
	}
	return views, nil
}
