package admin

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool, ytMock *testutil.YouTubeServiceMock) *Service {
	t.Helper()
	return &Service{
		db:        pool,
		ytService: ytMock,
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
	}
}

// seedValuableChannel はテスト用にValuableChannelを作成し、そのexternalChannelIDを返す
func seedValuableChannel(t *testing.T, pool *pgxpool.Pool, ytMock *testutil.YouTubeServiceMock, reason, description string) string {
	t.Helper()
	svc := newTestService(t, pool, ytMock)
	externalChannelID := "@testchannel"
	_, err := svc.CreateNewValuableChannel(context.Background(), externalChannelID, reason, description)
	if err != nil {
		t.Fatalf("seedValuableChannel: %v", err)
	}
	return externalChannelID
}
