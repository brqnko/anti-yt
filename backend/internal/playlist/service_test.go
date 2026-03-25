package playlist

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestService_CreatePlaylist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		title       string
		description string
		visibility  string
		plType      string
		wantErr     bool
	}{
		{
			name:        "success",
			title:       "My Playlist",
			description: "A great playlist",
			visibility:  "private",
			plType:      "normal",
		},
		{
			name:        "invalid_title_empty",
			title:       "",
			description: "desc",
			visibility:  "private",
			plType:      "normal",
			wantErr:     true,
		},
		{
			name:        "invalid_visibility",
			title:       "My Playlist",
			description: "desc",
			visibility:  "invalid",
			plType:      "normal",
			wantErr:     true,
		},
		{
			name:        "invalid_playlist_type",
			title:       "My Playlist",
			description: "desc",
			visibility:  "private",
			plType:      "invalid",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			svc := newTestService(t, pool, ytMock)
			pl, err := svc.CreatePlaylist(ctx, userID, tt.title, tt.description, tt.visibility, tt.plType, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if pl == nil {
				t.Fatal("playlist should not be nil")
			}
			if pl.Title.String() != tt.title {
				t.Fatalf("expected title %q, got %q", tt.title, pl.Title.String())
			}
		})
	}
}

func TestService_CreatePlaylist_WithImport(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	svc := newTestService(t, pool, ytMock)
	baseURL := "PLxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	pl, err := svc.CreatePlaylist(ctx, userID, "Imported", "desc", "private", "normal", &baseURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pl == nil {
		t.Fatal("playlist should not be nil")
	}
	if pl.VideoCount != 1 {
		t.Fatalf("expected video count 1, got %d", pl.VideoCount)
	}
}

func TestService_GetPlaylists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		seedCount   int
		limit       int32
		wantCount   int
		wantHasNext bool
	}{
		{name: "empty", seedCount: 0, limit: 10, wantCount: 0, wantHasNext: false},
		{name: "returns_playlists", seedCount: 2, limit: 10, wantCount: 2, wantHasNext: false},
		{name: "pagination", seedCount: 3, limit: 2, wantCount: 2, wantHasNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()
			svc := newTestService(t, pool, ytMock)

			for i := 0; i < tt.seedCount; i++ {
				seedPlaylist(t, pool, ytMock, userID)
			}

			playlists, hasNext, err := svc.GetPlaylists(ctx, userID, nil, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(playlists) != tt.wantCount {
				t.Fatalf("expected %d playlists, got %d", tt.wantCount, len(playlists))
			}
			if hasNext != tt.wantHasNext {
				t.Fatalf("expected hasNext=%v, got %v", tt.wantHasNext, hasNext)
			}
		})
	}
}

func TestService_GetPlaylistDetail(t *testing.T) {
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
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			var playlistID uuid.UUID
			if tt.seed {
				pl := seedPlaylist(t, pool, ytMock, userID)
				playlistID = pl.ID
			} else {
				playlistID = uuid.Must(uuid.NewV7())
			}

			svc := newTestService(t, pool, ytMock)
			view, err := svc.GetPlaylistDetail(ctx, userID, playlistID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if view.PlaylistId != playlistID {
				t.Fatalf("expected playlist ID %s, got %s", playlistID, view.PlaylistId)
			}
		})
	}
}

func TestService_DeletePlaylist(t *testing.T) {
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
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			var playlistID uuid.UUID
			if tt.seed {
				pl := seedPlaylist(t, pool, ytMock, userID)
				playlistID = pl.ID
			} else {
				playlistID = uuid.Must(uuid.NewV7())
			}

			svc := newTestService(t, pool, ytMock)
			err := svc.DeletePlaylist(ctx, userID, playlistID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 削除後は取得できないことを確認
			_, err = svc.GetPlaylistDetail(ctx, userID, playlistID)
			if err == nil {
				t.Fatal("expected error after deletion, got nil")
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				t.Fatalf("expected pgx.ErrNoRows, got %v", err)
			}
		})
	}
}

func TestService_UpdatePlaylist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		newTitle *string
		newDesc  *string
		wantErr  bool
	}{
		{name: "update_title", newTitle: strPtr("Updated Title")},
		{name: "update_description", newDesc: strPtr("Updated Description")},
		{name: "update_both", newTitle: strPtr("New Title"), newDesc: strPtr("New Desc")},
		{name: "no_changes"},
		{name: "invalid_title_empty", newTitle: strPtr(""), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			pl := seedPlaylist(t, pool, ytMock, userID)

			svc := newTestService(t, pool, ytMock)
			updated, err := svc.UpdatePlaylist(ctx, userID, pl.ID, tt.newTitle, tt.newDesc)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.newTitle != nil && updated.Title.String() != *tt.newTitle {
				t.Fatalf("expected title %q, got %q", *tt.newTitle, updated.Title.String())
			}
			if tt.newDesc != nil && updated.Description.String() != *tt.newDesc {
				t.Fatalf("expected description %q, got %q", *tt.newDesc, updated.Description.String())
			}
		})
	}
}

func TestService_UpdatePlaylist_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()
	svc := newTestService(t, pool, ytMock)

	title := "New Title"
	_, err := svc.UpdatePlaylist(ctx, userID, uuid.Must(uuid.NewV7()), &title, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestService_InsertVideoIntoPlaylist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	pl := seedPlaylist(t, pool, ytMock, userID)
	videoID := seedVideo(t, pool)

	svc := newTestService(t, pool, ytMock)
	err := svc.InsertVideoIntoPlaylist(ctx, userID, pl.ID, videoID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// プレイリストアイテムが1件あることを確認
	items, _, err := svc.GetPlaylistItems(ctx, userID, pl.ID, nil, 10)
	if err != nil {
		t.Fatalf("GetPlaylistItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestService_RemoveVideoFromPlaylist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	pl := seedPlaylist(t, pool, ytMock, userID)
	videoID := seedVideo(t, pool)

	svc := newTestService(t, pool, ytMock)

	// 先に追加
	if err := svc.InsertVideoIntoPlaylist(ctx, userID, pl.ID, videoID); err != nil {
		t.Fatalf("setup: InsertVideoIntoPlaylist failed: %v", err)
	}

	// 削除
	if err := svc.RemoveVideoFromPlaylist(ctx, userID, pl.ID, videoID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 0件になっていることを確認
	items, _, err := svc.GetPlaylistItems(ctx, userID, pl.ID, nil, 10)
	if err != nil {
		t.Fatalf("GetPlaylistItems failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items after removal, got %d", len(items))
	}
}

func TestService_GetPlaylistItems(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	pl := seedPlaylist(t, pool, ytMock, userID)

	svc := newTestService(t, pool, ytMock)

	// 空のプレイリスト
	items, hasNext, err := svc.GetPlaylistItems(ctx, userID, pl.ID, nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
	if hasNext {
		t.Fatal("expected hasNext=false")
	}
}

func strPtr(s string) *string { return &s }
