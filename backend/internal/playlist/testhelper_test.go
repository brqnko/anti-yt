package playlist

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool, ytMock *ServiceMock) *Service {
	t.Helper()
	return &Service{
		db:         pool,
		ytService:  ytMock,
		playlistQS: NewPlaylistQueryService(pool),
	}
}

func defaultYTMock() *ServiceMock {
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

	return &ServiceMock{
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
		FetchRSSFeedFunc: func(ctx context.Context, channelID youtube_d.ChannelID) ([]youtube_d.VideoID, error) {
			return []youtube_d.VideoID{"dQw4w9WgXcQ"}, nil
		},
	}
}

// seedPlaylist はプレイリストを作成し返す
func seedPlaylist(t *testing.T, pool *pgxpool.Pool, ytMock *ServiceMock, userID uuid.UUID) *Playlist {
	t.Helper()
	svc := newTestService(t, pool, ytMock)
	pl, err := svc.CreatePlaylist(context.Background(), userID, "Test Playlist", "A test playlist", "private", "normal", nil)
	if err != nil {
		t.Fatalf("seedPlaylist: %v", err)
	}
	return pl
}

// seedVideo はチャンネルと動画をDBに保存し動画のpublic_idを返す
func seedVideo(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	now := time.Now().UTC()

	if err := q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: "@testchannel",
		ExternalID:       "UCxxxxxxxxxxxxxxxxxxxxxx",
	}); err != nil {
		t.Fatalf("seedVideo: ClearStaleChannelCustomID: %v", err)
	}

	chRow, err := q.UpsertChannel(ctx, sqlc.UpsertChannelParams{
		ExternalID:                "UCxxxxxxxxxxxxxxxxxxxxxx",
		ExternalDisplayName:       "Test Channel",
		ExternalCustomID:          "@testchannel",
		ExternalIconUrl:           "https://example.com/icon.jpg",
		ExternalDescription:       "A test channel",
		ExternalSubscribersCount:  1000,
		ExternalCreatedAt:         time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalUploadsPlaylistID: "UUxxxxxxxxxxxxxxxxxxxxxx",
		PublicID:                  uuid.Must(uuid.NewV7()),
		RssFetchedAt:              now,
		FetchedAt:                 now,
	})
	if err != nil {
		t.Fatalf("seedVideo: UpsertChannel: %v", err)
	}

	videoPubID := uuid.Must(uuid.NewV7())
	if _, err := q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             chRow.PublicID,
		ExternalID:            "dQw4w9WgXcQ",
		ExternalTitle:         "Test Video",
		ExternalDescription:   "A test video",
		FetchedAt:             now,
		ExternalCreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
		ExternalLengthSeconds: 300,
		ID:                    videoPubID,
	}); err != nil {
		t.Fatalf("seedVideo: UpsertVideo: %v", err)
	}

	return videoPubID
}
