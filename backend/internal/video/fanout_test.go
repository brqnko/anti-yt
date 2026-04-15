package video_test

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/video"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFanOut(t *testing.T) {
	ctx := context.Background()

	t.Run("pushes video to all subscribers", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)

		channelPublicID := seedChannel(t, ctx, q)
		user1 := seedUserWithSuffix(t, ctx, q, "a")
		user2 := seedUserWithSuffix(t, ctx, q, "b")
		// 非購読者は fan-out に含まれないことを確認するため別ユーザーも作る
		user3 := seedUserWithSuffix(t, ctx, q, "c")

		subscribe(t, ctx, q, user1, channelPublicID)
		subscribe(t, ctx, q, user2, channelPublicID)

		v := seedVideo(t, ctx, q, channelPublicID)

		feedRepo := testutil.NewFakeFeedRepository()
		require.NoError(t, video.FanOut(ctx, q, feedRepo, v))

		assert.True(t, feedRepo.Has(user1, v.ID))
		assert.True(t, feedRepo.Has(user2, v.ID))
		assert.False(t, feedRepo.Has(user3, v.ID))
	})

	t.Run("no subscribers is a no-op", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)

		channelPublicID := seedChannel(t, ctx, q)
		v := seedVideo(t, ctx, q, channelPublicID)

		feedRepo := testutil.NewFakeFeedRepository()
		require.NoError(t, video.FanOut(ctx, q, feedRepo, v))
	})
}

func seedChannel(t *testing.T, ctx context.Context, q sqlc.Querier) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	_, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                "UC" + id.String()[:22],
		ExternalDisplayName:       "Test Channel",
		ExternalCustomID:          "@testchannel_" + id.String()[:8],
		ExternalIconUrl:           "https://example.com/icon.jpg",
		ExternalDescription:       "desc",
		ExternalSubscribersCount:  1000,
		ExternalCreatedAt:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalUploadsPlaylistID: "UU" + id.String()[:22],
		PublicID:                  id,
		RssFetchedAt:              time.Now(),
		FetchedAt:                 time.Now(),
		BulkFetchedAt:             time.Now(),
	})
	require.NoError(t, err)
	return id
}

func seedUserWithSuffix(t *testing.T, ctx context.Context, q sqlc.Querier, suffix string) uuid.UUID {
	t.Helper()
	userPublicID := uuid.Must(uuid.NewV7())
	authPublicID := uuid.Must(uuid.NewV7())
	_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         "https://accounts.google.com",
		Sub:            "test-sub-" + suffix + "-" + userPublicID.String(),
		LastLoggedInAt: time.Now(),
		PublicID:       authPublicID,
	})
	require.NoError(t, err)
	_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User " + suffix,
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    3600,
		JoinedAt:                  time.Now(),
		PublicID:                  userPublicID,
		UserAuthorizationPublicID: authPublicID,
	})
	require.NoError(t, err)
	return userPublicID
}

func subscribe(t *testing.T, ctx context.Context, q sqlc.Querier, userPublicID, channelPublicID uuid.UUID) {
	t.Helper()
	_, err := q.InsertSubscription(ctx, sqlc.InsertSubscriptionParams{
		UserPublicID: userPublicID,
		ChannelID:    channelPublicID,
		SubscribedAt: time.Now(),
	})
	require.NoError(t, err)
}

func seedVideo(t *testing.T, ctx context.Context, q sqlc.Querier, channelPublicID uuid.UUID) *video.Video {
	t.Helper()
	videoPublicID := uuid.Must(uuid.NewV7())
	_, err := q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             channelPublicID,
		ExternalID:            "vid" + videoPublicID.String()[:8],
		ExternalTitle:         "Test Video",
		ExternalDescription:   "desc",
		FetchedAt:             time.Now(),
		ExternalCreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
		ExternalLengthSeconds: 300,
		ID:                    videoPublicID,
	})
	require.NoError(t, err)
	return &video.Video{
		ID:        videoPublicID,
		ChannelID: channelPublicID,
	}
}
