package history

import (
	"context"
	"errors"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type HistoryRepository interface {
	GetLastHeartbeatForUpdate(ctx context.Context, userID uuid.UUID) (
		_ *Heartbeat,
		videoLengthSeconds int,
		heartbeatID int64,
		lastUpdatedAt time.Time,
		_ error,
	)
	CreateHeartbeat(ctx context.Context, heartbeat *Heartbeat) error
	UpdateHeartbeat(ctx context.Context, heartbeatID int64, heartbeat *Heartbeat) error
	MarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) error
	UnmarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) error
}

type historyRepositoryImpl struct {
	q sqlc.Querier
}

func NewHistoryRepository(q sqlc.Querier) HistoryRepository {
	return &historyRepositoryImpl{q: q}
}

func (h *historyRepositoryImpl) GetLastHeartbeatForUpdate(ctx context.Context, userID uuid.UUID) (
	_ *Heartbeat,
	videoLengthSeconds int,
	heartbeatID int64,
	lastUpdatedAt time.Time,
	err error,
) {
	defer util.Wrap(&err, "history.(*historyRepositoryImpl).GetLastHeartbeatForUpdate")

	heartbeat, err := h.q.GetLastHeartbeatForUpdate(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, 0, time.Time{}, core.ErrNotFound
		}
		return nil, 0, 0, time.Time{}, err
	}

	heartbeatDomain, err := NewHeartbeat(
		heartbeat.VideoID,
		userID,
		heartbeat.WatchPositionSeconds,
		WithHeartbeatWatchStartAt(heartbeat.WatchStartAt),
		WithHeartbeatWatchEndAt(heartbeat.WatchEndAt),
		WithHeartbeatID(heartbeat.PublicID),
	)
	if err != nil {
		return nil, 0, 0, time.Time{}, err
	}

	return heartbeatDomain, heartbeat.VideoLength, heartbeat.TVideoWatchID, heartbeat.UpdatedAt, nil
}

func (h *historyRepositoryImpl) CreateHeartbeat(ctx context.Context, heartbeat *Heartbeat) (err error) {
	defer util.Wrap(&err, "history.(*historyRepositoryImpl).CreateHeartbeat")

	if err := h.q.InsertHeartbeat(ctx, sqlc.InsertHeartbeatParams{
		UserPublicID:         heartbeat.UserID,
		VideoPublicID:        heartbeat.VideoID,
		PublicID:             heartbeat.ID,
		WatchStartAt:         heartbeat.WatchStartAt,
		WatchEndAt:           heartbeat.WatchEndAt,
		WatchPositionSeconds: int(heartbeat.WatchPositionSeconds),
	}); err != nil {
		return err
	}

	return nil
}

func (h *historyRepositoryImpl) UpdateHeartbeat(ctx context.Context, heartbeatID int64, heartbeat *Heartbeat) (err error) {
	defer util.Wrap(&err, "history.(*historyRepositoryImpl).UpdateHeartbeat")

	if err := h.q.UpdateHeartbeat(ctx, sqlc.UpdateHeartbeatParams{
		WatchPositionSeconds: int(heartbeat.WatchPositionSeconds),
		WatchStartAt:         heartbeat.WatchStartAt,
		WatchEndAt:           heartbeat.WatchEndAt,
		TVideoWatchID:        heartbeatID,
	}); err != nil {
		return err
	}

	return nil
}

func (h *historyRepositoryImpl) MarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "history.(*historyRepositoryImpl).MarkVideoWatched(userID=%s, videoID=%s)", userID, videoID)

	return h.q.MarkVideoWatched(ctx, sqlc.MarkVideoWatchedParams{
		UserPublicID:  userID,
		VideoPublicID: videoID,
	})
}

func (h *historyRepositoryImpl) UnmarkVideoWatched(ctx context.Context, userID, videoID uuid.UUID) (err error) {
	defer util.Wrap(&err, "history.(*historyRepositoryImpl).UnmarkVideoWatched(userID=%s, videoID=%s)", userID, videoID)

	return h.q.UnmarkVideoWatched(ctx, sqlc.UnmarkVideoWatchedParams{
		UserPublicID:  userID,
		VideoPublicID: videoID,
	})
}
