package history

import (
	"context"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

type HistoryRepository interface {
	Heartbeat(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, positionSeconds int, playlistID *uuid.UUID) error
	MarkVideoWatched(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID) error
	UnmarkVideoWatched(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID) error
	Import(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, watchStartAt, watchEndAt time.Time) error
}

type historyRepositoryImpl struct{}

func NewHistoryRepository() HistoryRepository {
	return &historyRepositoryImpl{}
}

func (h *historyRepositoryImpl) Heartbeat(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, positionSeconds int, playlistID *uuid.UUID) (err error) {
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

func (h *historyRepositoryImpl) MarkVideoWatched(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "historyRepository.MarkVideoWatched(userID=%s, videoID=%s)", userID, videoID)

	return q.MarkVideoWatched(ctx, sqlc.MarkVideoWatchedParams{
		UserPublicID:  userID,
		VideoPublicID: videoID,
	})
}

func (h *historyRepositoryImpl) UnmarkVideoWatched(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "historyRepository.UnmarkVideoWatched(userID=%s, videoID=%s)", userID, videoID)

	return q.UnmarkVideoWatched(ctx, sqlc.UnmarkVideoWatchedParams{
		UserPublicID:  userID,
		VideoPublicID: videoID,
	})
}

func (h *historyRepositoryImpl) Import(ctx context.Context, q sqlc.Querier, userID, videoID uuid.UUID, watchStartAt, watchEndAt time.Time) (err error) {
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
