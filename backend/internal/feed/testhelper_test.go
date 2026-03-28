package feed

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool, ytMock *testutil.YouTubeServiceMock) *Service {
	t.Helper()
	return &Service{
		db:               pool,
		ytService:        ytMock,
		feedQS:           NewFeedQueryService(pool),
		rssFetchDuration: 1 * time.Hour,
	}
}

func defaultYTMock() *testutil.YouTubeServiceMock {
	ch, _ := youtube_d.NewChannel(
		"UCxxxxxxxxxxxxxxxxxxxxxx",
		"Test Channel",
		"@testchannel",
		"A test channel",
		"https://example.com/icon.jpg",
		1000,
		"UUxxxxxxxxxxxxxxxxxxxxxx",
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	vid, _ := youtube_d.NewVideo(
		"dQw4w9WgXcQ",
		"UCxxxxxxxxxxxxxxxxxxxxxx",
		"Test Video",
		"A test video",
		"https://example.com/thumb.jpg",
		300,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	return &testutil.YouTubeServiceMock{
		FetchChannelDetailFunc: func(ctx context.Context, channelIDs []youtube_d.ChannelID) (map[youtube_d.ChannelID]youtube_d.Channel, error) {
			return map[youtube_d.ChannelID]youtube_d.Channel{
				"UCxxxxxxxxxxxxxxxxxxxxxx": ch,
			}, nil
		},
		FetchChannelDetailByIDOrHandleFunc: func(ctx context.Context, channelID string) (youtube_d.Channel, error) {
			return ch, nil
		},
		FetchPlaylistVideoIDsFunc: func(ctx context.Context, playlistID string, pageToken string) ([]youtube_d.VideoID, string, error) {
			return []youtube_d.VideoID{"dQw4w9WgXcQ"}, "", nil
		},
		FetchVideoDetailFunc: func(ctx context.Context, videoIDs []youtube_d.VideoID) (map[youtube_d.VideoID]youtube_d.Video, error) {
			return map[youtube_d.VideoID]youtube_d.Video{
				"dQw4w9WgXcQ": vid,
			}, nil
		},
		SearchVideoIDsFunc: func(ctx context.Context, query string, pageToken string, opts youtube_d.SearchOptions) ([]youtube_d.VideoID, string, error) {
			return []youtube_d.VideoID{"dQw4w9WgXcQ"}, "", nil
		},
	}
}

// seedSubscription はチャンネルを作成し、ユーザーを購読させる。チャンネルの public_id を返す。
// rss_fetched_at を古くして FindToFetchRSSForUpdate にヒットするようにする。
func seedSubscription(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	now := time.Now().UTC()
	channelPubID := uuid.Must(uuid.NewV7())

	if err := q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: "@testchannel",
		ExternalID:       "UCxxxxxxxxxxxxxxxxxxxxxx",
	}); err != nil {
		t.Fatalf("seedSubscription: ClearStaleChannelCustomID: %v", err)
	}

	if _, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                "UCxxxxxxxxxxxxxxxxxxxxxx",
		ExternalDisplayName:       "Test Channel",
		ExternalCustomID:          "@testchannel",
		ExternalIconUrl:           "https://example.com/icon.jpg",
		ExternalDescription:       "A test channel",
		ExternalSubscribersCount:  1000,
		ExternalCreatedAt:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalUploadsPlaylistID: "UUxxxxxxxxxxxxxxxxxxxxxx",
		PublicID:                  channelPubID,
		RssFetchedAt:              now.Add(-2 * time.Hour), // rssFetchDuration(1h) より古くする
		FetchedAt:                 now,
	}); err != nil {
		t.Fatalf("seedSubscription: UpsertChannel: %v", err)
	}

	if _, err := q.InsertSubscription(ctx, sqlc.InsertSubscriptionParams{
		UserPublicID: userID,
		ChannelID:    channelPubID,
		SubscribedAt: now,
	}); err != nil {
		t.Fatalf("seedSubscription: InsertSubscription: %v", err)
	}

	return channelPubID
}
