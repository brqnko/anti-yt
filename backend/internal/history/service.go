package history

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (s *Service) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int, playlistID *uuid.UUID, loc *time.Location) (_ *int, err error) {
	defer util.Wrap(&err, "Service.Heartbeat")

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to rollback", slog.Any("error", err))
		}
	}()
	q := sqlc.New(tx)

	if err := database_d.TryAdLock(ctx, q, userID[:]); err != nil {
		return nil, err
	}

	if err := NewHistoryRepository().Heartbeat(ctx, q, userID, videoID, positionSeconds, playlistID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
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
