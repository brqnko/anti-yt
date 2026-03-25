package video

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestService_GetVideoDetail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seed    bool
		wantErr bool
	}{
		{name: "success", seed: true},
		{name: "not_found", seed: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			var videoID uuid.UUID
			if tt.seed {
				videoID = seedChannelAndVideo(t, pool)
			} else {
				videoID = uuid.Must(uuid.NewV7())
			}

			svc := newTestService(t, pool)
			view, err := svc.GetVideoDetail(ctx, videoID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, pgx.ErrNoRows) {
					t.Fatalf("expected pgx.ErrNoRows, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if view.VideoId != videoID {
				t.Fatalf("expected video ID %s, got %s", videoID, view.VideoId)
			}
			if view.ExternalVideoTitle != "Test Video" {
				t.Fatalf("expected title %q, got %q", "Test Video", view.ExternalVideoTitle)
			}
			if view.ExternalChannelDisplayName != "Test Channel" {
				t.Fatalf("expected channel name %q, got %q", "Test Channel", view.ExternalChannelDisplayName)
			}
			if view.ExternalVideoId != "dQw4w9WgXcQ" {
				t.Fatalf("expected external video ID %q, got %q", "dQw4w9WgXcQ", view.ExternalVideoId)
			}
		})
	}
}
