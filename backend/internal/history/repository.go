package history

import (
	"context"
	"time"

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

	publicID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := h.q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userID,
		VideoPublicID:        videoID,
		WatchPositionSeconds: positionSeconds,
		PublicID:             publicID,
		WatchStartAt:         now,
		WatchEndAt:           now.Add(2 * time.Minute),
	}); err != nil {
		return err
	}
	return nil
}

var _ HistoryRepository = (*historyRepositoryImpl)(nil)
