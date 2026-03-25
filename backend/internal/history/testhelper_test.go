package history

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
		db:        pool,
		historyQS: NewHistoryQueryService(pool),
	}
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
