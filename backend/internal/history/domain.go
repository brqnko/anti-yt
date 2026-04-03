package history

import (
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

var farFuture = time.Date(9999, 12, 32, 0, 0, 0, 0, time.UTC)

var (
	ErrWatchPositionSecondsIsMinus = core.NewDomainError("watch_position_seconds_is_minus", "watch position seconds is minus")
)

type WatchPositionSeconds int

func NewWatchPositionSeconds(seconds int) (WatchPositionSeconds, error) {
	if seconds < 0 {
		return 0, ErrWatchPositionSecondsIsMinus
	}

	return WatchPositionSeconds(seconds), nil
}

type Heartbeat struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	VideoID              uuid.UUID
	WatchStartAt         time.Time
	WatchEndAt           time.Time
	WatchPositionSeconds WatchPositionSeconds
}

type HeartbeatOption func(h *Heartbeat)

func WithHeartbeatWatchStartAt(time time.Time) HeartbeatOption {
	return func(h *Heartbeat) {
		h.WatchStartAt = time
	}
}

func WithHeartbeatWatchEndAt(time time.Time) HeartbeatOption {
	return func(h *Heartbeat) {
		h.WatchEndAt = time
	}
}

func WithHeartbeatID(id uuid.UUID) HeartbeatOption {
	return func(h *Heartbeat) {
		h.ID = id
	}
}

func NewHeartbeat(videoID, userID uuid.UUID, watchPositionSeconds int, opts ...HeartbeatOption) (_ *Heartbeat, err error) {
	defer util.Wrap(&err, "history.NewHeartbeat")

	secs, err := NewWatchPositionSeconds(watchPositionSeconds)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	h := &Heartbeat{
		ID:                   id,
		VideoID:              videoID,
		WatchStartAt:         time.Now().UTC(),
		WatchEndAt:           farFuture,
		WatchPositionSeconds: secs,
		UserID:               userID,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}

func (h *Heartbeat) Rotate(videoID uuid.UUID, watchPositionSeconds, lastVideoLength int, lastUpdatedAt time.Time) (_ *Heartbeat, err error) {
	defer util.Wrap(&err, "history.(*Heartbeat).Rotate")

	// 違う動画を再生した場合
	if videoID != h.VideoID {
		h.WatchEndAt = time.Now().UTC()
		heartbeat, err := NewHeartbeat(videoID, h.UserID, watchPositionSeconds)
		if err != nil {
			return nil, err
		}
		return heartbeat, nil
	}

	// 最後の更新から2分以上経過していた場合はcloseして、新しく視聴を再開したとみなす
	if time.Now().UTC().Sub(lastUpdatedAt).Abs().Minutes() > 2 {
		h.WatchEndAt = lastUpdatedAt.Add(time.Minute)
		heartbeat, err := NewHeartbeat(videoID, h.UserID, watchPositionSeconds)
		if err != nil {
			return nil, err
		}
		return heartbeat, nil
	}

	// 動画を最後まで見終わった場合(30秒のpaddingつき)
	// YouTube動画でエンディングがあることを考慮した
	if lastVideoLength > 0 && watchPositionSeconds+30 >= lastVideoLength {
		h.WatchEndAt = time.Now().UTC()
		h.WatchPositionSeconds = WatchPositionSeconds(watchPositionSeconds)
		return nil, nil
	}

	// 同じ動画を継続視聴中: positionだけ更新
	h.WatchPositionSeconds = WatchPositionSeconds(watchPositionSeconds)
	return nil, nil
}
