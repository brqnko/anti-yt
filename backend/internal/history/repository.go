package history

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

type HistoryRepository struct{}

func NewHistoryRepository() *HistoryRepository {
	return &HistoryRepository{}
}

func (h *HistoryRepository) Heartbeat(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, positionSeconds int, playlistID *uuid.UUID) (err error) {
	defer util.Wrap(&err, "historyRepository.Heartbeat(userID=%s, videoID=%s)", userID, videoID)

	publicID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	if err := q.CloseStaleWatchSessions(ctx, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	if err := q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userID,
		VideoPublicID:        videoID,
		WatchPositionSeconds: positionSeconds,
		PublicID:             publicID,
		WatchStartAt:         now,
	}); err != nil {
		return err
	}

	if playlistID != nil {
		if err := q.PushRecentPlaylistId(ctx, sqlc.PushRecentPlaylistIdParams{
			PlaylistPublicID: *playlistID,
			UserPublicID:     userID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryRepository) Import(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, watchStartAt, watchEndAt time.Time) (err error) {
	defer util.Wrap(&err, "historyRepository.Import(userID=%s, videoID=%s)", userID, videoID)

	publicID, err := uuid.NewV7()
	if err != nil {
		return err
	}

	if err := q.UpsertWatchHeartbeat(ctx, sqlc.UpsertWatchHeartbeatParams{
		UserPublicID:         userID,
		VideoPublicID:        videoID,
		WatchPositionSeconds: 0,
		PublicID:             publicID,
		WatchStartAt:         watchStartAt,
	}); err != nil {
		return err
	}
	return nil
}
