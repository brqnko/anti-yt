package history

import (
	"context"
	"fmt"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) (*Service, error) {
	return &Service{
		db: db,
	}, nil
}

func (s *Service) GetHistory(ctx context.Context, limit int, cursor *uuid.UUID) (items []HistoryItem, hasNext bool, err error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return nil, false, err
	}

	q := sqlc.New(s.db)

	rows, err := q.GetHistory(ctx, sqlc.GetHistoryParams{
		UserID:     userID,
		Cursor:     cursor,
		QueryLimit: int32(limit + 1),
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to getHistory: %w", err)
	}

	if len(rows) > limit {
		hasNext = true
		rows = rows[:limit]
	}

	items = make([]HistoryItem, len(rows))
	for i, row := range rows {
		item, err := NewHistoryItem(
			row.VideoID,
			row.ExternalVideoID,
			row.ExternalVideoTitle,
			row.ExternalVideoThumbnailUrl,
			row.ExternalVideoLengthSeconds,
			row.ExternalVideoCreatedAt,
			row.WatchPositionSeconds,
			row.WatchedAt,
			row.ChannelID,
			row.ExternalChannelID,
			row.ExternalChannelDisplayName,
			row.ExternalChannelIconUrl,
		)
		if err != nil {
			return nil, false, fmt.Errorf("failed to newHistoryItem: %w", err)
		}
		items[i] = item
	}

	return items, hasNext, nil
}

func (s *Service) GetStatisticsByWeek(ctx context.Context, targetWeek time.Time) (WeeklyStatistics, error) {
	userID, err := util.UserIDFromContext(ctx)
	if err != nil {
		return WeeklyStatistics{}, err
	}

	rows, err := sqlc.New(s.db).GetUserStatisticsByWeek(ctx, sqlc.GetUserStatisticsByWeekParams{
		UserID:    userID,
		StartDate: targetWeek,
		EndDate:   targetWeek.Add(7 * 24 * time.Hour), // NOTE: postgresql側で+ '7 days'するとsqlcがパースエラー起こす
	})
	if err != nil {
		return WeeklyStatistics{}, fmt.Errorf("failed to getUserStatisticsByWeek: %w", err)
	}

	daily := make([]DailyStatistics, len(rows))
	for i, row := range rows {
		daily[i] = NewDailyStatistics(row.WatchDate.Time, row.VideoCount, row.WatchSum)
	}

	return NewWeeklyStatistics(targetWeek, daily, ""), nil
}
