package history

import (
	"context"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestService_Heartbeat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		screenLimit     *int // nil = unlimited
		wantRemainingNil bool
	}{
		{
			name:            "unlimited_screen_time",
			screenLimit:     nil,
			wantRemainingNil: true,
		},
		{
			name:            "with_screen_limit",
			screenLimit:     intPtr(3600),
			wantRemainingNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUserWithScreenLimit(t, pool, tt.screenLimit)
			videoID := seedVideo(t, pool)

			svc := newTestService(t, pool)
			remaining, err := svc.Heartbeat(ctx, userID, videoID, 30)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantRemainingNil {
				if remaining != nil {
					t.Fatalf("expected remaining to be nil (unlimited), got %d", *remaining)
				}
			} else {
				if remaining == nil {
					t.Fatal("expected remaining to be non-nil")
				}
				if *remaining < 0 {
					t.Fatalf("remaining should be >= 0, got %d", *remaining)
				}
			}
		})
	}
}

func TestService_GetHistory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		heartbeats  int
		limit       int
		wantCount   int
		wantHasNext bool
	}{
		{name: "empty", heartbeats: 0, limit: 10, wantCount: 0, wantHasNext: false},
		{name: "with_history", heartbeats: 1, limit: 10, wantCount: 1, wantHasNext: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			videoID := seedVideo(t, pool)

			svc := newTestService(t, pool)

			for i := 0; i < tt.heartbeats; i++ {
				if _, err := svc.Heartbeat(ctx, userID, videoID, 30); err != nil {
					t.Fatalf("setup: Heartbeat failed: %v", err)
				}
			}

			views, hasNext, err := svc.GetHistory(ctx, userID, tt.limit, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(views) != tt.wantCount {
				t.Fatalf("expected %d history items, got %d", tt.wantCount, len(views))
			}
			if hasNext != tt.wantHasNext {
				t.Fatalf("expected hasNext=%v, got %v", tt.wantHasNext, hasNext)
			}
		})
	}
}

func TestService_GetStatisticsByWeek(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	videoID := seedVideo(t, pool)

	svc := newTestService(t, pool)

	// heartbeat を送信して視聴記録を作る
	if _, err := svc.Heartbeat(ctx, userID, videoID, 60); err != nil {
		t.Fatalf("setup: Heartbeat failed: %v", err)
	}

	// watch_end_at を確実に過去にクローズ（ListDailyWatchStatsByRange は watch_end_at <= CURRENT_TIMESTAMP を条件にしている）
	_, err := pool.Exec(ctx, "UPDATE t_video_watch SET watch_end_at = CURRENT_TIMESTAMP - interval '1 second'")
	if err != nil {
		t.Fatalf("setup: close sessions failed: %v", err)
	}

	// 今日を含む範囲で統計を取得
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	views, err := svc.GetStatisticsByWeek(ctx, userID, today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// heartbeat を送ったので少なくとも1日分のデータがあるはず
	if len(views) == 0 {
		t.Fatal("expected at least 1 day of statistics")
	}
}

func TestService_GetStatisticsByWeek_Empty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)

	svc := newTestService(t, pool)

	// 遠い過去の週を指定（データなし）
	pastWeek := time.Date(2000, 1, 3, 0, 0, 0, 0, time.UTC)
	views, err := svc.GetStatisticsByWeek(ctx, userID, pastWeek)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("expected 0 statistics, got %d", len(views))
	}
}

func TestService_Heartbeat_InvalidVideo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)

	svc := newTestService(t, pool)
	_, err := svc.Heartbeat(ctx, userID, uuid.Must(uuid.NewV7()), 30)
	if err == nil {
		t.Fatal("expected error for non-existent video, got nil")
	}
}

func intPtr(v int) *int { return &v }
