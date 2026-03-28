package feed

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestService_GetFeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		subscribe   bool
		limit       int32
		wantErr     bool
	}{
		{name: "empty_no_subscription", subscribe: false, limit: 10},
		{name: "with_subscription", subscribe: true, limit: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			if tt.subscribe {
				seedSubscription(t, pool, userID)
			}

			svc := newTestService(t, pool, ytMock)
			videos, hasNext, err := svc.GetFeed(ctx, userID, nil, tt.limit)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.subscribe {
				// RSS フェッチで動画が保存されるので1件以上
				if len(videos) == 0 {
					t.Fatal("expected at least 1 video in feed")
				}
			} else {
				if len(videos) != 0 {
					t.Fatalf("expected 0 videos, got %d", len(videos))
				}
			}
			_ = hasNext
		})
	}
}

func TestService_GetFeed_PlaylistFetchError(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	seedSubscription(t, pool, userID)

	ytMock.FetchPlaylistVideoIDsFunc = func(ctx context.Context, playlistID string, pageToken string) ([]youtube_d.VideoID, string, error) {
		return nil, "", errors.New("playlist fetch error")
	}

	svc := newTestService(t, pool, ytMock)
	_, _, err := svc.GetFeed(ctx, userID, nil, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Search(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		limit      int
		resultIDs  []youtube_d.VideoID
		nextToken  string
		wantCount  int
		wantNext   bool
		wantCursor bool
	}{
		{
			name:      "success",
			query:     "test",
			limit:     10,
			resultIDs: []youtube_d.VideoID{"dQw4w9WgXcQ"},
			wantCount: 1,
			wantNext:  false,
		},
		{
			name:      "empty_results",
			query:     "nothing",
			limit:     10,
			resultIDs: nil,
			wantCount: 0,
			wantNext:  false,
		},
		{
			name:       "with_pagination",
			query:      "test",
			limit:      10,
			resultIDs:  []youtube_d.VideoID{"dQw4w9WgXcQ"},
			nextToken:  "next-page-token",
			wantCount:  1,
			wantNext:   true,
			wantCursor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			ytMock := defaultYTMock()

			ytMock.SearchVideoIDsFunc = func(ctx context.Context, query string, pageToken string, opts youtube_d.SearchOptions) ([]youtube_d.VideoID, string, error) {
				return tt.resultIDs, tt.nextToken, nil
			}

			svc := newTestService(t, pool, ytMock)
			items, hasNext, cursor, err := svc.Search(ctx, tt.query, tt.limit, nil, youtube_d.SearchOptions{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != tt.wantCount {
				t.Fatalf("expected %d items, got %d", tt.wantCount, len(items))
			}
			if hasNext != tt.wantNext {
				t.Fatalf("expected hasNext=%v, got %v", tt.wantNext, hasNext)
			}
			if tt.wantCursor {
				if cursor == nil {
					t.Fatal("expected cursor to be non-nil")
				}
			} else {
				if cursor != nil {
					t.Fatalf("expected cursor to be nil, got %q", *cursor)
				}
			}
		})
	}
}

func TestService_Search_Error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	ytMock := defaultYTMock()

	ytMock.SearchVideoIDsFunc = func(ctx context.Context, query string, pageToken string, opts youtube_d.SearchOptions) ([]youtube_d.VideoID, string, error) {
		return nil, "", errors.New("search api error")
	}

	svc := newTestService(t, pool, ytMock)
	_, _, _, err := svc.Search(ctx, "test", 10, nil, youtube_d.SearchOptions{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_Search_WithCursor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	ytMock := defaultYTMock()

	var receivedPageToken string
	ytMock.SearchVideoIDsFunc = func(ctx context.Context, query string, pageToken string, opts youtube_d.SearchOptions) ([]youtube_d.VideoID, string, error) {
		receivedPageToken = pageToken
		return []youtube_d.VideoID{"dQw4w9WgXcQ"}, "", nil
	}

	svc := newTestService(t, pool, ytMock)
	cursor := "page2"
	_, _, _, err := svc.Search(ctx, "test", 10, &cursor, youtube_d.SearchOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedPageToken != "page2" {
		t.Fatalf("expected pageToken %q, got %q", "page2", receivedPageToken)
	}
}

func TestService_GetFeed_NoUserSubscriptions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := uuid.Must(uuid.NewV7()) // ユーザーが存在しなくてもエラーにならないはず
	ytMock := defaultYTMock()

	svc := newTestService(t, pool, ytMock)
	videos, _, err := svc.GetFeed(ctx, userID, nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 0 {
		t.Fatalf("expected 0 videos, got %d", len(videos))
	}
}
