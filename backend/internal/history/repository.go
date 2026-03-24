package history

import (
	"context"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
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

func (h *historyRepositoryImpl) Heartbeat(ctx context.Context, userID, videoID uuid.UUID, positionSeconds int) (err error) {
	defer util.Wrap(&err, "historyRepository.Heartbeat(userID=%s, videoID=%s)", userID, videoID)
	if err := h.q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userID,
		VideoPublicID:        videoID,
		WatchPositionSeconds: positionSeconds,
	}); err != nil {
		return err
	}
	return nil
}

var _ HistoryRepository = (*historyRepositoryImpl)(nil)
