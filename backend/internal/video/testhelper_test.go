package video

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newTestService(t *testing.T, pool *pgxpool.Pool) *Service {
	t.Helper()
	return &Service{
		videoQS: NewVideoQueryService(pool),
	}
}

// seedChannelAndVideo はチャンネルと動画をDBに保存し、動画のpublic_idを返す
func seedChannelAndVideo(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	q := sqlc.New(pool)
	now := time.Now().UTC()
	channelPubID := uuid.Must(uuid.NewV7())

	// ClearStaleChannelCustomID → UpsertChannel の順でチャンネルを保存
	if err := q.ClearStaleChannelCustomID(ctx, sqlc.ClearStaleChannelCustomIDParams{
		ExternalCustomID: "@testchannel",
		ExternalID:       "UCxxxxxxxxxxxxxxxxxxxxxx",
	}); err != nil {
		t.Fatalf("seedChannelAndVideo: ClearStaleChannelCustomID: %v", err)
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
		PublicID:                  channelPubID,
		RssFetchedAt:              now,
		FetchedAt:                 now,
	})
	if err != nil {
		t.Fatalf("seedChannelAndVideo: UpsertChannel: %v", err)
	}

	// 動画を保存
	videoPubID := uuid.Must(uuid.NewV7())
	if _, err := q.UpsertVideo(ctx, sqlc.UpsertVideoParams{
		ChannelID:             chRow.PublicID,
		ExternalID:            "dQw4w9WgXcQ",
		ExternalTitle:         "Test Video",
		ExternalDescription:   "A test video description",
		FetchedAt:             now,
		ExternalCreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		ExternalThumbnailUrl:  "https://example.com/thumb.jpg",
		ExternalLengthSeconds: 300,
		ID:                    videoPubID,
	}); err != nil {
		t.Fatalf("seedChannelAndVideo: UpsertVideo: %v", err)
	}

	return videoPubID
}
