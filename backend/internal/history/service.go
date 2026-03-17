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

type HistoryItem struct {
	VideoID                    uuid.UUID
	ExternalVideoID            string
	ExternalVideoTitle         string
	ExternalVideoThumbnailURL  string
	ExternalVideoLengthSeconds int
	WatchPositionSeconds       int
	WatchedAt                  time.Time
	ChannelID                  uuid.UUID
	ExternalChannelID          string
	ExternalChannelDisplayName string
	ExternalChannelIconURL     string
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
		return nil, false, fmt.Errorf("getHistory: %w", err)
	}

	if len(rows) > limit {
		hasNext = true
		rows = rows[:limit]
	}

	items = make([]HistoryItem, len(rows))
	for i, row := range rows {
		items[i] = HistoryItem{
			VideoID:                    row.VideoID,
			ExternalVideoID:            row.ExternalVideoID,
			ExternalVideoTitle:         row.ExternalVideoTitle,
			ExternalVideoThumbnailURL:  row.ExternalVideoThumbnailUrl,
			ExternalVideoLengthSeconds: row.ExternalVideoLengthSeconds,
			WatchPositionSeconds:       row.WatchPositionSeconds,
			WatchedAt:                  row.WatchedAt,
			ChannelID:                  row.ChannelID,
			ExternalChannelID:          row.ExternalChannelID,
			ExternalChannelDisplayName: row.ExternalChannelDisplayName,
			ExternalChannelIconURL:     row.ExternalChannelIconUrl,
		}
	}

	return items, hasNext, nil
}
