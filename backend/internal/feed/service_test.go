package feed_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/feed"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetFeed(t *testing.T) {
	ctx := context.Background()

	t.Run("empty feed when no subscriptions", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		feedRepo := testutil.NewFeedRepo(t, sqlc.New(db))
		svc := feed.NewService(db, new(ClientMock{}), feedRepo)

		videos, hasMore, err := svc.GetFeed(ctx, userPublicID, nil, 10)

		require.NoError(t, err)
		assert.Empty(t, videos)
		assert.False(t, hasMore)
	})

	t.Run("returns videos from subscribed channel", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		feedRepo := testutil.NewFeedRepo(t, sqlc.New(db))
		chSvc := channel.NewService(db, ytMock, feedRepo)

		_, err := chSvc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		// SubscribeChannel 直後に PushMany で subscriber feed に反映されているはず
		feedItems, err := feedRepo.Get(ctx, userPublicID, nil, math.MaxInt64)
		require.NoError(t, err)
		assert.Len(t, feedItems, 1)

		feedSvc := feed.NewService(db, ytMock, feedRepo)

		videos, hasMore, err := feedSvc.GetFeed(ctx, userPublicID, nil, 10)

		require.NoError(t, err)
		assert.Len(t, videos, 1)
		assert.False(t, hasMore)
		assert.Equal(t, "Test Video", videos[0].ExternalVideoTitle)
	})

	t.Run("hasMore true when more videos than limit", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		vid1 := newTestVideoWith(t, "aaaAAAAAAAAA", "Video 1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		vid2 := newTestVideoWith(t, "bbbBBBBBBBBB", "Video 2", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
		videoMap := map[youtube_d.VideoID]youtube_d.Video{vid1.ID: vid1, vid2.ID: vid2}

		ytMock := newYTMock(testCh, videoMap)
		feedRepo := testutil.NewFeedRepo(t, sqlc.New(db))
		chSvc := channel.NewService(db, ytMock, feedRepo)

		_, err := chSvc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		feedSvc := feed.NewService(db, ytMock, feedRepo)

		videos, hasMore, err := feedSvc.GetFeed(ctx, userPublicID, nil, 1)

		require.NoError(t, err)
		assert.Len(t, videos, 1)
		assert.True(t, hasMore)
	})
}

func TestService_Search(t *testing.T) {
	ctx := context.Background()

	t.Run("empty results", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return []youtube_d.SearchItem{}, "", nil
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		items, hasNext, nextCursor, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		require.NoError(t, err)
		assert.Empty(t, items)
		assert.False(t, hasNext)
		assert.Nil(t, nextCursor)
	})

	t.Run("returns video search results", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return []youtube_d.SearchItem{{Type: youtube_d.SearchItemTypeVideo, VideoID: testVid.ID}}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
			FetchChannelDetailFunc: func(_ context.Context, _ []youtube_d.ChannelID) (map[youtube_d.ChannelID]youtube_d.Channel, error) {
				return map[youtube_d.ChannelID]youtube_d.Channel{testCh.ID: testCh}, nil
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		items, hasNext, nextCursor, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "video", items[0].Type)
		assert.Equal(t, "Test Video", items[0].ExternalVideoTitle)
		assert.Equal(t, "Test Channel", items[0].ExternalChannelDisplayName)
		assert.False(t, hasNext)
		assert.Nil(t, nextCursor)
	})

	t.Run("hasNext true with nextPageToken", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return []youtube_d.SearchItem{{Type: youtube_d.SearchItemTypeVideo, VideoID: testVid.ID}}, "next-page-token", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
			FetchChannelDetailFunc: func(_ context.Context, _ []youtube_d.ChannelID) (map[youtube_d.ChannelID]youtube_d.Channel, error) {
				return map[youtube_d.ChannelID]youtube_d.Channel{testCh.ID: testCh}, nil
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		items, hasNext, nextCursor, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.True(t, hasNext)
		require.NotNil(t, nextCursor)
		assert.Equal(t, "next-page-token", *nextCursor)
	})

	t.Run("cursor is passed as pageToken", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		var capturedPageToken string
		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, pageToken string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				capturedPageToken = pageToken
				return []youtube_d.SearchItem{{Type: youtube_d.SearchItemTypeVideo, VideoID: testVid.ID}}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
			FetchChannelDetailFunc: func(_ context.Context, _ []youtube_d.ChannelID) (map[youtube_d.ChannelID]youtube_d.Channel, error) {
				return map[youtube_d.ChannelID]youtube_d.Channel{testCh.ID: testCh}, nil
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		cursor := "my-page-token"
		_, _, _, err := svc.Search(ctx, "test query", 10, &cursor, youtube_d.SearchOptions{})

		require.NoError(t, err)
		assert.Equal(t, "my-page-token", capturedPageToken)
	})

	t.Run("SearchIDs error returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		searchErr := errors.New("search failed")
		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return nil, "", searchErr
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		_, _, _, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		assert.ErrorIs(t, err, searchErr)
	})

	t.Run("FetchVideoDetail error returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		detailErr := errors.New("video detail failed")
		testVid := newTestVideo(t)

		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return []youtube_d.SearchItem{{Type: youtube_d.SearchItemTypeVideo, VideoID: testVid.ID}}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return nil, detailErr
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		_, _, _, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		assert.ErrorIs(t, err, detailErr)
	})

	t.Run("FetchChannelDetail error returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		channelErr := errors.New("channel detail failed")
		testVid := newTestVideo(t)

		svc := feed.NewService(db, new(ClientMock{
			SearchIDsFunc: func(_ context.Context, _ string, _ string, _ youtube_d.SearchOptions) ([]youtube_d.SearchItem, string, error) {
				return []youtube_d.SearchItem{{Type: youtube_d.SearchItemTypeVideo, VideoID: testVid.ID}}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
			FetchChannelDetailFunc: func(_ context.Context, _ []youtube_d.ChannelID) (map[youtube_d.ChannelID]youtube_d.Channel, error) {
				return nil, channelErr
			},
		}), testutil.NewFeedRepo(t, sqlc.New(db)))

		_, _, _, err := svc.Search(ctx, "test query", 10, nil, youtube_d.SearchOptions{})

		assert.ErrorIs(t, err, channelErr)
	})
}

func newTestChannel(t *testing.T) youtube_d.Channel {
	t.Helper()
	ch, err := youtube_d.NewChannel(
		"UCxxxxxxxxxxxxxxxxxxxxxx",
		"Test Channel",
		"@testchannel",
		"A test channel",
		"https://example.com/icon.jpg",
		1000,
		"UUxxxxxxxxxxxxxxxxxxxxxx",
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)
	return ch
}

func newTestVideo(t *testing.T) youtube_d.Video {
	t.Helper()
	return newTestVideoWith(t, "dQw4w9WgXcQ", "Test Video", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
}

func newTestVideoWith(t *testing.T, id, title string, createdAt time.Time) youtube_d.Video {
	t.Helper()
	v, err := youtube_d.NewVideo(
		id,
		"UCxxxxxxxxxxxxxxxxxxxxxx",
		title,
		"A test video",
		"https://example.com/thumb.jpg",
		300,
		createdAt,
	)
	require.NoError(t, err)
	return v
}

func newYTMock(ch youtube_d.Channel, videos map[youtube_d.VideoID]youtube_d.Video) *ClientMock {
	ids := make([]youtube_d.VideoID, 0, len(videos))
	for id := range videos {
		ids = append(ids, id)
	}
	return new(ClientMock{
		FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
			return ch, nil
		},
		FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
			return ids, "", nil
		},
		FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
			return videos, nil
		},
	})
}

func setupUser(t *testing.T, ctx context.Context, q sqlc.Querier) uuid.UUID {
	t.Helper()
	userPublicID := uuid.Must(uuid.NewV7())
	authPublicID := uuid.Must(uuid.NewV7())
	_, err := q.SaveAuthorization(ctx, sqlc.SaveAuthorizationParams{
		Issuer:         "https://accounts.google.com",
		Sub:            "test-sub-" + userPublicID.String(),
		LastLoggedInAt: time.Now(),
		PublicID:       authPublicID,
	})
	require.NoError(t, err)
	_, err = q.InsertUser(ctx, sqlc.InsertUserParams{
		DisplayName:               "Test User",
		LanguageCode:              "ja",
		DailyScreenTimeSeconds:    3600,
		JoinedAt:                  time.Now(),
		PublicID:                  userPublicID,
		UserAuthorizationPublicID: authPublicID,
	})
	require.NoError(t, err)
	return userPublicID
}
