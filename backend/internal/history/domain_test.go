package history_test

import (
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewWatchPositionSeconds(t *testing.T) {
	t.Parallel()

	type arg struct {
		seconds int
	}

	type want struct {
		seconds history.WatchPositionSeconds
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{seconds: 120},
			want:    &want{seconds: history.WatchPositionSeconds(120)},
			wantErr: nil,
		},
		"zero": {
			arg:     arg{seconds: 0},
			want:    &want{seconds: history.WatchPositionSeconds(0)},
			wantErr: nil,
		},
		"minus": {
			arg:     arg{seconds: -1},
			want:    nil,
			wantErr: history.ErrWatchPositionSecondsIsMinus,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := history.NewWatchPositionSeconds(c.arg.seconds)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.seconds, got)
			}
		})
	}
}

func TestWatchPositionSeconds_IsFinished(t *testing.T) {
	t.Parallel()

	type arg struct {
		watchPosition   history.WatchPositionSeconds
		videoLengthSecs int
	}

	cases := map[string]struct {
		arg  arg
		want bool
	}{
		"finished": {
			arg:  arg{watchPosition: 90, videoLengthSecs: 100},
			want: true,
		},
		"not finished": {
			arg:  arg{watchPosition: 10, videoLengthSecs: 100},
			want: false,
		},
		"exactly at boundary": {
			arg:  arg{watchPosition: 70, videoLengthSecs: 100},
			want: true,
		},
		"video length is zero": {
			arg:  arg{watchPosition: 90, videoLengthSecs: 0},
			want: false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got := c.arg.watchPosition.IsFinished(c.arg.videoLengthSecs)

			// assert
			assert.Equal(t, c.want, got)
		})
	}
}

func TestNewHeartbeat(t *testing.T) {
	t.Parallel()

	videoID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	type arg struct {
		videoID              uuid.UUID
		userID               uuid.UUID
		watchPositionSeconds int
	}

	type want struct {
		videoID              uuid.UUID
		userID               uuid.UUID
		watchPositionSeconds history.WatchPositionSeconds
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:     arg{videoID: videoID, userID: userID, watchPositionSeconds: 120},
			want:    &want{videoID: videoID, userID: userID, watchPositionSeconds: 120},
			wantErr: nil,
		},
		"minus watch position": {
			arg:     arg{videoID: videoID, userID: userID, watchPositionSeconds: -1},
			want:    nil,
			wantErr: history.ErrWatchPositionSecondsIsMinus,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := history.NewHeartbeat(c.arg.videoID, c.arg.userID, c.arg.watchPositionSeconds)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.videoID, got.VideoID)
				assert.Equal(t, c.want.userID, got.UserID)
				assert.Equal(t, c.want.watchPositionSeconds, got.WatchPositionSeconds)
			}
		})
	}
}

func TestHeartbeat_Rotate_StaleDifferentVideo_ClosesAtLastUpdatedAt(t *testing.T) {
	t.Parallel()

	videoID := uuid.Must(uuid.NewV7())
	otherVideoID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	h, err := history.NewHeartbeat(videoID, userID, 0)
	assert.NoError(t, err)

	lastUpdatedAt := time.Now().UTC().Add(-3 * time.Hour)
	got, result, err := h.Rotate(otherVideoID, 30, 200, lastUpdatedAt)
	assert.NoError(t, err)
	assert.Equal(t, history.RotateResultContinue, result)
	assert.NotNil(t, got)

	expectedEnd := lastUpdatedAt.Add(time.Minute)
	assert.WithinDuration(t, expectedEnd, h.WatchEndAt, time.Second,
		"stale session should close at lastUpdatedAt+1min, not at time.Now()")
}

func TestHeartbeat_Rotate(t *testing.T) {
	t.Parallel()

	videoID := uuid.Must(uuid.NewV7())
	otherVideoID := uuid.Must(uuid.NewV7())
	userID := uuid.Must(uuid.NewV7())

	active := func() *history.Heartbeat {
		h, _ := history.NewHeartbeat(videoID, userID, 0)
		return h
	}
	closed := func() *history.Heartbeat {
		h, _ := history.NewHeartbeat(videoID, userID, 0, history.WithHeartbeatWatchEndAt(time.Now().UTC().Add(-time.Hour)))
		return h
	}

	type arg struct {
		heartbeat            func() *history.Heartbeat
		videoID              uuid.UUID
		watchPositionSeconds int
		lastVideoLength      int
		lastUpdatedAt        time.Time
	}

	type want struct {
		newHeartbeat   bool // 新しいheartbeatが返るか
		originalClosed bool // 元のheartbeatがcloseされるか
		watchPosition  history.WatchPositionSeconds
		result         history.RotateResult
	}

	cases := map[string]struct {
		arg     arg
		want    want
		wantErr error
	}{
		"already closed + same video": {
			arg:  arg{heartbeat: closed, videoID: videoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			want: want{newHeartbeat: false, originalClosed: true, result: history.RotateResultIgnored},
		},
		"already closed + different video": {
			arg:  arg{heartbeat: closed, videoID: otherVideoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			want: want{newHeartbeat: true, originalClosed: true, result: history.RotateResultContinue},
		},
		"active + different video": {
			arg:  arg{heartbeat: active, videoID: otherVideoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			want: want{newHeartbeat: true, originalClosed: true, result: history.RotateResultContinue},
		},
		"active + same video + last updated over 5 min": {
			arg:  arg{heartbeat: active, videoID: videoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC().Add(-6 * time.Minute)},
			want: want{newHeartbeat: true, originalClosed: true, result: history.RotateResultContinue},
		},
		"active + different video + last updated over 5 min": {
			arg:  arg{heartbeat: active, videoID: otherVideoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC().Add(-10 * time.Minute)},
			want: want{newHeartbeat: true, originalClosed: true, result: history.RotateResultContinue},
		},
		"active + same video + last updated 3 min ago (within threshold)": {
			arg:  arg{heartbeat: active, videoID: videoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC().Add(-3 * time.Minute)},
			want: want{newHeartbeat: false, originalClosed: false, watchPosition: 30, result: history.RotateResultContinue},
		},
		"active + same video + finished": {
			arg:  arg{heartbeat: active, videoID: videoID, watchPositionSeconds: 180, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			want: want{newHeartbeat: false, originalClosed: true, watchPosition: 180, result: history.RotateResultFinished},
		},
		"active + same video + continuing": {
			arg:  arg{heartbeat: active, videoID: videoID, watchPositionSeconds: 30, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			want: want{newHeartbeat: false, originalClosed: false, watchPosition: 30, result: history.RotateResultContinue},
		},
		"invalid watch position": {
			arg:     arg{heartbeat: active, videoID: videoID, watchPositionSeconds: -1, lastVideoLength: 200, lastUpdatedAt: time.Now().UTC()},
			wantErr: history.ErrWatchPositionSecondsIsMinus,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// arrange
			h := c.arg.heartbeat()

			// act
			got, result, err := h.Rotate(c.arg.videoID, c.arg.watchPositionSeconds, c.arg.lastVideoLength, c.arg.lastUpdatedAt)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.result, result)
				assert.Equal(t, c.want.newHeartbeat, got != nil)
				assert.Equal(t, c.want.originalClosed, h.WatchEndAt.Before(time.Now().UTC()))
				if c.want.watchPosition != 0 {
					assert.Equal(t, c.want.watchPosition, h.WatchPositionSeconds)
				}
			}
		})
	}
}
