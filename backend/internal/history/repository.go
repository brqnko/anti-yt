package history

import (
	"context"
	"fmt"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
)

type HistoryRepository interface {
	Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int) error
}

type historyRepositoryImpl struct {
	q sqlc.Querier
}

func NewHistoryRepository(q sqlc.Querier) HistoryRepository {
	return &historyRepositoryImpl{
		q: q,
	}
}

func (h *historyRepositoryImpl) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int) error {
	if err := h.q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userID,
		VideoPublicID:        videoID,
		WatchPositionSeconds: positionSeconds,
	}); err != nil {
		return fmt.Errorf("failed to heartbeat(historyRepository.Heartbeat): %w", err)
	}
	return nil
}

var _ HistoryRepository = (*historyRepositoryImpl)(nil)
