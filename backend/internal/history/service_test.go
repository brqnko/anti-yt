package history_test

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

// setupChannel はテスト用チャンネルを作成し publicID を返す
func setupChannel(t *testing.T, ctx context.Context, q sqlc.Querier) uuid.UUID {
	t.Helper()
	channelPublicID := uuid.Must(uuid.NewV7())
	_, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                "UC" + channelPublicID.String()[:24],
		ExternalDisplayName:       "Test Channel",
		ExternalCustomID:          "@tc" + channelPublicID.String()[:8],
		ExternalIconUrl:           "https://example.com/icon.jpg",
		ExternalDescription:       "A test channel",
		ExternalSubscribersCount:  1000,
		ExternalCreatedAt:         time.Now().Add(-365 * 24 * time.Hour),
		ExternalUploadsPlaylistID: "UU" + channelPublicID.String()[:24],
		PublicID:                  channelPublicID,
		RssFetchedAt:             time.Now(),
		FetchedAt:                time.Now(),
		BulkFetchedAt:            time.Now(),
	})
	require.NoError(t, err)
	return channelPublicID
}

// setupVideo はテスト用動画を作成し publicID を返す。lengthSeconds で動画の長さを指定できる。
func setupVideo(t *testing.T, ctx context.Context, q sqlc.Querier, channelPublicID uuid.UUID, lengthSeconds int) uuid.UUID {
	t.Helper()
	videoPublicID := uuid.Must(uuid.NewV7())
	_, err := q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             channelPublicID,
		ExternalID:            videoPublicID.String()[:16],
		ExternalTitle:         "Test Video",
		ExternalDescription:   "A test video",
		FetchedAt:             time.Now(),
		ExternalCreatedAt:     time.Now().Add(-24 * time.Hour),
		ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
		ExternalLengthSeconds: lengthSeconds,
		ID:                    videoPublicID,
	})
	require.NoError(t, err)
	return videoPublicID
}

func TestService_Heartbeat(t *testing.T) {
	ctx := context.Background()
	loc := time.UTC

	t.Run("first heartbeat creates new entry", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		remaining, err := svc.Heartbeat(ctx, userID, videoID, 10, nil, loc)

		require.NoError(t, err)
		require.NotNil(t, remaining)
		assert.Greater(t, *remaining, 0)
	})

	t.Run("continuing same video updates position", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		_, err := svc.Heartbeat(ctx, userID, videoID, 10, nil, loc)
		require.NoError(t, err)

		remaining, err := svc.Heartbeat(ctx, userID, videoID, 30, nil, loc)

		require.NoError(t, err)
		require.NotNil(t, remaining)
	})

	t.Run("switching to different video", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		video1 := setupVideo(t, ctx, q, channelID, 300)
		video2 := setupVideo(t, ctx, q, channelID, 600)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		_, err := svc.Heartbeat(ctx, userID, video1, 10, nil, loc)
		require.NoError(t, err)

		remaining, err := svc.Heartbeat(ctx, userID, video2, 0, nil, loc)

		require.NoError(t, err)
		require.NotNil(t, remaining)
	})

	t.Run("with playlist id", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)
		// playlistIDが存在しなくてもPushRecentPlaylistIdはエラーにならない（UPDATEのみ）
		playlistID := uuid.Must(uuid.NewV7())

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		remaining, err := svc.Heartbeat(ctx, userID, videoID, 10, &playlistID, loc)

		require.NoError(t, err)
		require.NotNil(t, remaining)
	})

	t.Run("remaining is nil when unlimited", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)

		// daily_screen_time_seconds を 24*60*60（無制限）に設定
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
			DisplayName:               "Unlimited User",
			LanguageCode:              "ja",
			DailyScreenTimeSeconds:    24 * 60 * 60,
			JoinedAt:                  time.Now(),
			PublicID:                  userPublicID,
			UserAuthorizationPublicID: authPublicID,
		})
		require.NoError(t, err)

		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		remaining, err := svc.Heartbeat(ctx, userPublicID, videoID, 10, nil, loc)

		require.NoError(t, err)
		assert.Nil(t, remaining)
	})

	t.Run("nonexistent user returns error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		_, err := svc.Heartbeat(ctx, uuid.Must(uuid.NewV7()), videoID, 10, nil, loc)

		assert.Error(t, err)
	})
}

func TestService_MarkVideoWatched(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		err := svc.MarkVideoWatched(ctx, userID, videoID)

		require.NoError(t, err)
	})

	t.Run("mark same video twice is idempotent", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		err := svc.MarkVideoWatched(ctx, userID, videoID)
		require.NoError(t, err)

		err = svc.MarkVideoWatched(ctx, userID, videoID)
		require.NoError(t, err)
	})
}

func TestService_UnmarkVideoWatched(t *testing.T) {
	ctx := context.Background()

	t.Run("unmark after mark succeeds", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		err := svc.MarkVideoWatched(ctx, userID, videoID)
		require.NoError(t, err)

		err = svc.UnmarkVideoWatched(ctx, userID, videoID)
		require.NoError(t, err)
	})

	t.Run("unmark without mark does not error", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		err := svc.UnmarkVideoWatched(ctx, userID, videoID)
		require.NoError(t, err)
	})
}

func TestService_GetHistory(t *testing.T) {
	ctx := context.Background()
	loc := time.UTC

	t.Run("no history returns empty", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		views, hasMore, err := svc.GetHistory(ctx, userID, 10, nil, loc)

		require.NoError(t, err)
		assert.Empty(t, views)
		assert.False(t, hasMore)
	})

	t.Run("returns history after heartbeat", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		_, err := svc.Heartbeat(ctx, userID, videoID, 10, nil, loc)
		require.NoError(t, err)

		views, hasMore, err := svc.GetHistory(ctx, userID, 10, nil, loc)

		require.NoError(t, err)
		assert.Len(t, views, 1)
		assert.False(t, hasMore)
		assert.Equal(t, videoID, views[0].VideoId)
		assert.Equal(t, "Test Video", views[0].ExternalVideoTitle)
	})

	t.Run("pagination with hasMore", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		// 3つの動画のheartbeatを作成
		for range 3 {
			vid := setupVideo(t, ctx, q, channelID, 300)
			_, err := svc.Heartbeat(ctx, userID, vid, 10, nil, loc)
			require.NoError(t, err)
			// 別動画への切り替えを確実にするため、前の動画を閉じる
		}

		views, hasMore, err := svc.GetHistory(ctx, userID, 2, nil, loc)

		require.NoError(t, err)
		assert.Len(t, views, 2)
		assert.True(t, hasMore)
	})

	t.Run("cursor pagination", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		for range 3 {
			vid := setupVideo(t, ctx, q, channelID, 300)
			_, err := svc.Heartbeat(ctx, userID, vid, 10, nil, loc)
			require.NoError(t, err)
		}

		// 最初のページ
		firstPage, hasMore, err := svc.GetHistory(ctx, userID, 2, nil, loc)
		require.NoError(t, err)
		assert.Len(t, firstPage, 2)
		assert.True(t, hasMore)

		// 2ページ目（カーソル使用）
		cursor := firstPage[len(firstPage)-1].WatchId
		secondPage, hasMore, err := svc.GetHistory(ctx, userID, 2, &cursor, loc)
		require.NoError(t, err)
		assert.Len(t, secondPage, 1)
		assert.False(t, hasMore)
	})

	t.Run("nonexistent user returns empty", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		views, hasMore, err := svc.GetHistory(ctx, uuid.Must(uuid.NewV7()), 10, nil, loc)

		require.NoError(t, err)
		assert.Empty(t, views)
		assert.False(t, hasMore)
	})
}

func TestService_GetStatisticsByWeek(t *testing.T) {
	ctx := context.Background()
	loc := time.UTC

	t.Run("no data returns empty views", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		targetWeek := time.Now().In(loc).Truncate(24 * time.Hour)
		aiSummary, views, err := svc.GetStatisticsByWeek(ctx, userID, targetWeek, loc)

		require.NoError(t, err)
		assert.Nil(t, aiSummary)
		assert.Empty(t, views)
	})

	t.Run("returns statistics after heartbeat", func(t *testing.T) {
		db := testutil.NewTestPool(t)
		q := sqlc.New(db)
		userID := setupUser(t, ctx, q)
		channelID := setupChannel(t, ctx, q)
		videoID := setupVideo(t, ctx, q, channelID, 300)

		svc := history.NewService(db, testutil.NewFakeFeedRepository(sqlc.New(db)))

		// heartbeatを作成してから、完了させる（watch_end_atをcloseする）
		_, err := svc.Heartbeat(ctx, userID, videoID, 10, nil, loc)
		require.NoError(t, err)

		// 違う動画に切り替えてheartbeatをcloseさせる
		video2 := setupVideo(t, ctx, q, channelID, 300)
		_, err = svc.Heartbeat(ctx, userID, video2, 0, nil, loc)
		require.NoError(t, err)

		now := time.Now().In(loc)
		targetWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		_, views, err := svc.GetStatisticsByWeek(ctx, userID, targetWeek, loc)

		require.NoError(t, err)
		// closedされたheartbeatのみがstatisticsに含まれる
		assert.NotEmpty(t, views)
	})
}
