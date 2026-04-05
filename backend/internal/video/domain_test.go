package video_test

import (
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewVideo(t *testing.T) {
	t.Parallel()

	channelID := uuid.Must(uuid.NewV7())
	fetchedAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ytVideo := youtube_d.Video{
		ID:            "videoid123",
		ChannelID:     "channelid123",
		Title:         "test video",
		Description:   "test description",
		ThumbnailURL:  "https://example.com/thumbnail.jpg",
		LengthSeconds: 300,
		CreatedAt:     time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	type arg struct {
		channelID uuid.UUID
		fetchedAt time.Time
		video     youtube_d.Video
	}

	type want struct {
		channelID uuid.UUID
		fetchedAt time.Time
		video     youtube_d.Video
	}

	cases := map[string]struct {
		arg     arg
		want    *want
		wantErr error
	}{
		"success": {
			arg:  arg{channelID: channelID, fetchedAt: fetchedAt, video: ytVideo},
			want: &want{channelID: channelID, fetchedAt: fetchedAt, video: ytVideo},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// act
			got, err := video.NewVideo(c.arg.channelID, c.arg.fetchedAt, c.arg.video)

			// assert
			if c.wantErr != nil {
				assert.ErrorIs(t, err, c.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.want.channelID, got.ChannelID)
				assert.Equal(t, c.want.fetchedAt, got.FetchedAt)
				assert.Equal(t, c.want.video, got.Video)
			}
		})
	}
}
