package video_test

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetVideoDetail(t *testing.T) {
	ctx := context.Background()

	t.Run("non-existent video returns not found", func(t *testing.T) {
		db := testutil.NewTestPool(t)

		_, err := video.NewService(db).GetVideoDetail(ctx, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))

		assert.ErrorIs(t, err, core.ErrNotFound)
	})

	t.Run("existing video returns detail", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)

		// チャンネルを作成
		channelPublicID := uuid.Must(uuid.NewV7())
		_, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
			ExternalID:                "UCxxxxxxxxxxxxxxxxxxxxxx",
			ExternalDisplayName:       "Test Channel",
			ExternalCustomID:          "@testchannel",
			ExternalIconUrl:           "https://example.com/icon.jpg",
			ExternalDescription:       "A test channel",
			ExternalSubscribersCount:  1000,
			ExternalCreatedAt:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			ExternalUploadsPlaylistID: "UUxxxxxxxxxxxxxxxxxxxxxx",
			PublicID:                  channelPublicID,
			RssFetchedAt:             time.Now(),
			FetchedAt:                time.Now(),
			BulkFetchedAt:            time.Now(),
			LastSeenAt:               time.Now(),
		})
		require.NoError(t, err)

		// 動画を作成
		videoPublicID := uuid.Must(uuid.NewV7())
		videoCreatedAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err = q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
			ChannelID:             channelPublicID,
			ExternalID:            "dQw4w9WgXcQ",
			ExternalTitle:         "Test Video",
			ExternalDescription:   "A test video",
			FetchedAt:             time.Now(),
			ExternalCreatedAt:     videoCreatedAt,
			ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
			ExternalLengthSeconds: 300,
			ID:                    videoPublicID,
		})
		require.NoError(t, err)

		// action
		detail, err := video.NewService(db).GetVideoDetail(ctx, uuid.Must(uuid.NewV7()), videoPublicID)

		// assert
		require.NoError(t, err)
		assert.Equal(t, videoPublicID, detail.VideoId)
		assert.Equal(t, "dQw4w9WgXcQ", detail.ExternalVideoId)
		assert.Equal(t, "Test Video", detail.ExternalVideoTitle)
		assert.Equal(t, "A test video", detail.ExternalVideoDescription)
		assert.Equal(t, "https://example.com/thumb.jpg", detail.ExternalVideoThumbnailUrl)
		assert.True(t, videoCreatedAt.Equal(detail.ExternalVideoCreatedAt))
		assert.Equal(t, channelPublicID, detail.ChannelId)
		assert.Equal(t, "@testchannel", detail.ChannelCustomId)
		assert.Equal(t, "Test Channel", detail.ExternalChannelDisplayName)
		assert.Equal(t, "https://example.com/icon.jpg", detail.ExternalChannelIconUrl)
		assert.Equal(t, uint64(1000), detail.ExternalChannelSubscribersCount)
		assert.False(t, detail.IsWatched)
		assert.False(t, detail.IsInWatchLater)
	})

}
