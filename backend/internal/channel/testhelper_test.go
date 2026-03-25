package channel

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool, ytMock *ServiceMock) *Service {
	t.Helper()
	return &Service{
		db:                pool,
		ytService:         ytMock,
		rssFetchDuration:  1 * time.Hour,
		channelQS:         NewChannelQueryService(pool),
		valuableChannelQS: NewValuableChannelQueryService(pool),
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

// seedSubscription はユーザーをチャンネルに登録し、チャンネルを返す
func seedSubscription(t *testing.T, pool *pgxpool.Pool, ytMock *ServiceMock, userID uuid.UUID) *Channel {
	t.Helper()
	svc := newTestService(t, pool, ytMock)
	ch, err := svc.SubscribeChannel(context.Background(), userID, "@testchannel")
	if err != nil {
		t.Fatalf("seedSubscription: %v", err)
	}
	return ch
}
