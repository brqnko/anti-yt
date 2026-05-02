package channel_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetSubscriptions(t *testing.T) {
	ctx := context.Background()

	t.Run("limit too small returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, _, err := svc.GetSubscriptions(ctx, uuid.New(), 0, nil)

		assert.ErrorIs(t, err, channel.ErrInvalidSubscriptionLimit)
	})

	t.Run("limit too large returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, _, err := svc.GetSubscriptions(ctx, uuid.New(), 51, nil)

		assert.ErrorIs(t, err, channel.ErrInvalidSubscriptionLimit)
	})

	t.Run("valid limit returns empty list", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := channel.NewService(db, &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		channels, hasNext, err := svc.GetSubscriptions(ctx, uuid.Must(uuid.NewV7()), 10, nil)

		require.NoError(t, err)
		assert.Empty(t, channels)
		assert.False(t, hasNext)
	})

	t.Run("returns subscribed channel after subscribe", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 24*time.Hour)

		_, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		channels, hasNext, err := svc.GetSubscriptions(ctx, userPublicID, 10, nil)

		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.False(t, hasNext)
		assert.Equal(t, testCh.DisplayName, channels[0].ExternalChannelDisplayName)
	})

	t.Run("hasNext true when more subscriptions than limit", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		// 2つの異なるチャンネルを登録する
		ch1, err := youtube_d.NewChannel(
			"UCaaaaaaaaaaaaaaaaaaaaa",
			"Channel A", "@channela", "desc", "https://example.com/a.jpg",
			100, "UUaaaaaaaaaaaaaaaaaaaaa",
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		)
		require.NoError(t, err)
		ch2, err := youtube_d.NewChannel(
			"UCbbbbbbbbbbbbbbbbbbbbb",
			"Channel B", "@channelb", "desc", "https://example.com/b.jpg",
			200, "UUbbbbbbbbbbbbbbbbbbbbb",
			time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		)
		require.NoError(t, err)

		callCount := 0
		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				callCount++
				if callCount == 1 {
					return ch1, nil
				}
				return ch2, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{}, nil
			},
		}, testutil.NewFakeFeedRepository(), 24*time.Hour)

		_, err = svc.SubscribeChannel(ctx, userPublicID, "@channela")
		require.NoError(t, err)
		_, err = svc.SubscribeChannel(ctx, userPublicID, "@channelb")
		require.NoError(t, err)

		// limit=1 で2件あれば hasNext=true
		channels, hasNext, err := svc.GetSubscriptions(ctx, userPublicID, 1, nil)

		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.True(t, hasNext)
	})
}

func TestService_GetChannelUploads(t *testing.T) {
	ctx := context.Background()

	t.Run("limit too small returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, _, err := svc.GetChannelUploads(ctx, uuid.New(), uuid.New(), nil, 0, "newer")

		assert.ErrorIs(t, err, channel.ErrInvalidGetUploadLimit)
	})

	t.Run("limit too large returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, _, err := svc.GetChannelUploads(ctx, uuid.New(), uuid.New(), nil, 51, "newer")

		assert.ErrorIs(t, err, channel.ErrInvalidGetUploadLimit)
	})

	t.Run("returns videos without RSS re-fetch", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})

		// rssFetchDuration を十分長くして RSS 再取得が発動しないようにする
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 24*time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		// SubscribeChannel で FetchPlaylistVideoIDs が1回呼ばれているのでリセット前の回数を記録
		callsBefore := len(ytMock.FetchPlaylistVideoIDsCalls())

		videos, hasNext, err := svc.GetChannelUploads(ctx, userPublicID, ch.ID, nil, 10, "newer")

		require.NoError(t, err)
		assert.Len(t, videos, 1)
		assert.False(t, hasNext)
		assert.Equal(t, "Test Video", videos[0].ExternalVideoTitle)
		// RSS 再取得が発動していないことを確認
		assert.Equal(t, callsBefore, len(ytMock.FetchPlaylistVideoIDsCalls()))
	})

	t.Run("triggers RSS re-fetch when stale", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		newVid := newTestVideoWith(t, "xyzABCDE123", "New Video After RSS", time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC))

		fetchCount := 0
		ytMock := &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				fetchCount++
				if fetchCount <= 1 {
					// SubscribeChannel 時は元の動画のみ
					return []youtube_d.VideoID{testVid.ID}, "", nil
				}
				// GetChannelUploads での RSS 再取得時は新しい動画も含む
				return []youtube_d.VideoID{testVid.ID, newVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, ids []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				result := map[youtube_d.VideoID]youtube_d.Video{}
				all := map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid, newVid.ID: newVid}
				for _, id := range ids {
					if v, ok := all[id]; ok {
						result[id] = v
					}
				}
				return result, nil
			},
		}

		// rssFetchDuration=0 で常に RSS 再取得が発動するようにする
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 0)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		videos, hasNext, err := svc.GetChannelUploads(ctx, userPublicID, ch.ID, nil, 10, "newer")

		require.NoError(t, err)
		assert.Len(t, videos, 2)
		assert.False(t, hasNext)
		// FetchPlaylistVideoIDs が2回呼ばれている(Subscribe時 + GetChannelUploads時)
		assert.Equal(t, 2, len(ytMock.FetchPlaylistVideoIDsCalls()))
	})

	t.Run("hasNext true when more videos than limit", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)

		// limit+1 以上の動画を用意して hasNext=true を確認する
		// limit=1 で 2本の動画があれば hasNext=true
		vid1 := newTestVideoWith(t, "aaaAAAAAAAA", "Video 1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		vid2 := newTestVideoWith(t, "bbbBBBBBBBB", "Video 2", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
		videoMap := map[youtube_d.VideoID]youtube_d.Video{vid1.ID: vid1, vid2.ID: vid2}

		ytMock := newYTMock(testCh, videoMap)
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 24*time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		videos, hasNext, err := svc.GetChannelUploads(ctx, userPublicID, ch.ID, nil, 1, "newer")

		require.NoError(t, err)
		assert.Len(t, videos, 1)
		assert.True(t, hasNext)
	})

	t.Run("hasNext false when videos fit in limit", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		vid1 := newTestVideoWith(t, "aaaAAAAAAAA", "Video 1", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		vid2 := newTestVideoWith(t, "bbbBBBBBBBB", "Video 2", time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC))
		videoMap := map[youtube_d.VideoID]youtube_d.Video{vid1.ID: vid1, vid2.ID: vid2}

		ytMock := newYTMock(testCh, videoMap)
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 24*time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		videos, hasNext, err := svc.GetChannelUploads(ctx, userPublicID, ch.ID, nil, 10, "newer")

		require.NoError(t, err)
		assert.Len(t, videos, 2)
		assert.False(t, hasNext)
	})

	t.Run("non-existent channel returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		svc := channel.NewService(db, &ServiceMock{}, testutil.NewFakeFeedRepository(), 24*time.Hour)

		_, _, err := svc.GetChannelUploads(ctx, userPublicID, uuid.Must(uuid.NewV7()), nil, 10, "newer")

		assert.Error(t, err)
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

// newYTMock は SubscribeChannel に必要な YouTube モックを返す
func newYTMock(ch youtube_d.Channel, videos map[youtube_d.VideoID]youtube_d.Video) *ServiceMock {
	ids := make([]youtube_d.VideoID, 0, len(videos))
	for id := range videos {
		ids = append(ids, id)
	}
	return &ServiceMock{
		FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
			return ch, nil
		},
		FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
			return ids, "", nil
		},
		FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
			return videos, nil
		},
	}
}

// setupUser はテスト用ユーザーを作成し publicID を返す
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

func TestService_SubscribeChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid channel text returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.SubscribeChannel(ctx, uuid.New(), "invalid")

		assert.Error(t, err)
	})

	t.Run("youtube fetch error returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		fetchErr := errors.New("youtube api error")
		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return youtube_d.Channel{}, fetchErr
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.SubscribeChannel(ctx, uuid.Must(uuid.NewV7()), "@testchannel")

		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("success with new channel", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")

		require.NoError(t, err)
		assert.Equal(t, testCh.DisplayName, ch.Channel.DisplayName)
		assert.Equal(t, testCh.ID, ch.Channel.ID)
	})

	t.Run("subscribe same channel twice by same user", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		// 2回目は既にチャンネルが保存されているのでfetchは不要だが、subscriptionのuniqueで失敗する可能性
		_, err = svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		assert.Error(t, err)
	})
}

func TestService_UnsubscribeChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("unsubscribe non-existent subscription returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := channel.NewService(db, &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		err := svc.UnsubscribeChannel(ctx, uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()))

		assert.Error(t, err)
	})

	t.Run("subscribe then unsubscribe succeeds", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		feedRepo := testutil.NewFakeFeedRepository()
		svc := channel.NewService(db, ytMock, feedRepo, time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)
		// Subscribe 後はチャンネルの既存動画が PushMany で feed に反映されている
		assert.Equal(t, 1, feedRepo.Count(userPublicID))

		err = svc.UnsubscribeChannel(ctx, userPublicID, ch.ID)
		assert.NoError(t, err)
		// Unsubscribe 後は DeleteMany で feed から削除されている
		assert.Equal(t, 0, feedRepo.Count(userPublicID))
	})
}

func TestService_GetChannelDetail(t *testing.T) {
	ctx := context.Background()

	t.Run("non-existent channel returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := channel.NewService(db, &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.GetChannelDetail(ctx, uuid.Must(uuid.NewV7()), util.Base64UUID(uuid.Must(uuid.NewV7())).String())

		assert.Error(t, err)
	})

	t.Run("existing channel returns detail", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		userPublicID := setupUser(t, ctx, sqlc.New(db))

		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), time.Hour)

		ch, err := svc.SubscribeChannel(ctx, userPublicID, "@testchannel")
		require.NoError(t, err)

		detail, err := svc.GetChannelDetail(ctx, userPublicID, util.Base64UUID(ch.ID).String())

		require.NoError(t, err)
		assert.Equal(t, testCh.DisplayName, detail.DisplayName)
		assert.Equal(t, testCh.CustomID, detail.CustomID)
	})
}

func TestService_GetChannelFeeds(t *testing.T) {
	ctx := context.Background()

	t.Run("empty valuable channels", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := channel.NewService(db, &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		channels, err := svc.GetChannelFeeds(ctx)

		require.NoError(t, err)
		assert.Empty(t, channels)
	})

	t.Run("returns valuable channel after creation", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)
		ytMock := newYTMock(testCh, map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid})
		svc := channel.NewService(db, ytMock, testutil.NewFakeFeedRepository(), 24*time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")
		require.NoError(t, err)

		channels, err := svc.GetChannelFeeds(ctx)

		require.NoError(t, err)
		assert.Len(t, channels, 1)
		assert.Equal(t, testCh.DisplayName, channels[0].ExternalChannelDisplayName)
		assert.Equal(t, "good channel", channels[0].ValuableDescription)
	})
}

func TestService_CreateNewValuableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid channel text returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "invalid", "education", "good channel")

		assert.Error(t, err)
	})

	t.Run("youtube fetch error returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		fetchErr := errors.New("youtube api error")
		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return youtube_d.Channel{}, fetchErr
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")

		assert.ErrorIs(t, err, fetchErr)
	})

	t.Run("success creates valuable channel", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{testVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		vc, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")

		require.NoError(t, err)
		assert.Equal(t, "education", vc.ValuableReasonCode.String())
		assert.Equal(t, "good channel", vc.ValuableDescription.String())
	})

	t.Run("invalid reason code returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{testVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "invalid_reason", "good channel")

		assert.Error(t, err)
	})
}

func TestService_RemoveValuableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid channel text returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		err := svc.RemoveValuableChannel(ctx, "invalid")

		assert.Error(t, err)
	})

	t.Run("create then remove valuable channel", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{testVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")
		require.NoError(t, err)

		err = svc.RemoveValuableChannel(ctx, "@testchannel")
		assert.NoError(t, err)
	})
}

func TestService_UpdateValuableChannel(t *testing.T) {
	ctx := context.Background()

	t.Run("invalid channel text returns error", func(t *testing.T) {
		svc := channel.NewService((*pgxpool.Pool)(nil), &ServiceMock{}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.UpdateValuableChannel(ctx, "invalid", nil, nil)

		assert.Error(t, err)
	})

	t.Run("create then update valuable channel", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{testVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")
		require.NoError(t, err)

		newReason := "technology"
		newDesc := "updated description"
		vc, err := svc.UpdateValuableChannel(ctx, "@testchannel", &newReason, &newDesc)

		require.NoError(t, err)
		assert.Equal(t, "technology", vc.ValuableReasonCode.String())
		assert.Equal(t, "updated description", vc.ValuableDescription.String())
	})

	t.Run("partial update only reason", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		testCh := newTestChannel(t)
		testVid := newTestVideo(t)

		svc := channel.NewService(db, &ServiceMock{
			FetchChannelDetailByIDOrHandleFunc: func(_ context.Context, _ string) (youtube_d.Channel, error) {
				return testCh, nil
			},
			FetchPlaylistVideoIDsFunc: func(_ context.Context, _ string, _ string) ([]youtube_d.VideoID, string, error) {
				return []youtube_d.VideoID{testVid.ID}, "", nil
			},
			FetchVideoDetailFunc: func(_ context.Context, _ []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
				return map[youtube_d.VideoID]youtube_d.Video{testVid.ID: testVid}, nil
			},
		}, testutil.NewFakeFeedRepository(), time.Hour)

		_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "good channel")
		require.NoError(t, err)

		newReason := "music"
		vc, err := svc.UpdateValuableChannel(ctx, "@testchannel", &newReason, nil)

		require.NoError(t, err)
		assert.Equal(t, "music", vc.ValuableReasonCode.String())
		assert.Equal(t, "good channel", vc.ValuableDescription.String())
	})
}
