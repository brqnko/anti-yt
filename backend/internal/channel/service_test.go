package channel

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestService_SubscribeChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		channelText string
		fetchErr    error
		wantErr     bool
	}{
		{
			name:        "success_by_handle",
			channelText: "@testchannel",
		},
		{
			name:        "success_by_channel_id",
			channelText: "UCxxxxxxxxxxxxxxxxxxxxxx",
		},
		{
			name:        "invalid_channel_text",
			channelText: "invalid",
			wantErr:     true,
		},
		{
			name:        "youtube_fetch_error",
			channelText: "@testchannel",
			fetchErr:    errors.New("youtube API error"),
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
			if tt.fetchErr != nil {
				ytMock.FetchChannelDetailByIDOrHandleFunc = func(ctx context.Context, channelID string) (youtube_d.Channel, error) {
					return youtube_d.Channel{}, tt.fetchErr
				}
			}

			svc := newTestService(t, pool, ytMock)
			ch, err := svc.SubscribeChannel(ctx, userID, tt.channelText)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ch == nil {
				t.Fatal("channel should not be nil")
			}
			if ch.Channel.DisplayName != "Test Channel" {
				t.Fatalf("expected display name %q, got %q", "Test Channel", ch.Channel.DisplayName)
			}
		})
	}
}

func TestService_SubscribeChannel_ExistingChannel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()

	svc := newTestService(t, pool, ytMock)

	// 1回目: チャンネル取得+登録
	_, err := svc.SubscribeChannel(ctx, userID, "@testchannel")
	if err != nil {
		t.Fatalf("first subscribe failed: %v", err)
	}

	// 2回目: 同じユーザーが同じチャンネル → duplicate subscription エラーの可能性
	user2ID := testutil.SeedUser(t, pool)
	ch, err := svc.SubscribeChannel(ctx, user2ID, "@testchannel")
	if err != nil {
		t.Fatalf("second user subscribe failed: %v", err)
	}
	// 既存チャンネルを使うので FetchChannelDetailByIDOrHandle は1回だけ呼ばれるべき
	if len(ytMock.FetchChannelDetailByIDOrHandleCalls()) != 1 {
		t.Fatalf("expected 1 fetch call, got %d", len(ytMock.FetchChannelDetailByIDOrHandleCalls()))
	}
	if ch == nil {
		t.Fatal("channel should not be nil")
	}
}

func TestService_UnsubscribeChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seed    bool
		wantErr bool
	}{
		{name: "success", seed: true, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			var channelID uuid.UUID
			if tt.seed {
				ch := seedSubscription(t, pool, ytMock, userID)
				channelID = ch.ID
			} else {
				channelID = uuid.Must(uuid.NewV7())
			}

			svc := newTestService(t, pool, ytMock)
			err := svc.UnsubscribeChannel(ctx, userID, channelID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 解除後はサブスクリプション一覧が空になることを確認
			subs, _, err := svc.GetSubscriptions(ctx, userID, 10, nil)
			if err != nil {
				t.Fatalf("GetSubscriptions failed: %v", err)
			}
			if len(subs) != 0 {
				t.Fatalf("expected 0 subscriptions after unsubscribe, got %d", len(subs))
			}
		})
	}
}

func TestService_GetSubscriptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		subCount    int
		limit       int32
		wantCount   int
		wantHasNext bool
		wantErr     bool
	}{
		{name: "empty", subCount: 0, limit: 10, wantCount: 0, wantHasNext: false},
		{name: "returns_subscriptions", subCount: 1, limit: 10, wantCount: 1, wantHasNext: false},
		{name: "invalid_limit_zero", subCount: 0, limit: 0, wantErr: true},
		{name: "invalid_limit_over", subCount: 0, limit: 51, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			if tt.subCount > 0 {
				seedSubscription(t, pool, ytMock, userID)
			}

			svc := newTestService(t, pool, ytMock)
			subs, hasNext, err := svc.GetSubscriptions(ctx, userID, tt.limit, nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(subs) != tt.wantCount {
				t.Fatalf("expected %d subscriptions, got %d", tt.wantCount, len(subs))
			}
			if hasNext != tt.wantHasNext {
				t.Fatalf("expected hasNext=%v, got %v", tt.wantHasNext, hasNext)
			}
		})
	}
}

func TestService_GetChannelUploads(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limit   int32
		wantErr bool
	}{
		{name: "success", limit: 10},
		{name: "invalid_limit_zero", limit: 0, wantErr: true},
		{name: "invalid_limit_over", limit: 51, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			ch := seedSubscription(t, pool, ytMock, userID)

			svc := newTestService(t, pool, ytMock)
			videos, _, err := svc.GetChannelUploads(ctx, userID, ch.ID, nil, tt.limit, "")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if videos == nil {
				t.Fatal("videos should not be nil")
			}
		})
	}
}

func TestService_GetChannelUploads_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	userID := testutil.SeedUser(t, pool)
	ytMock := defaultYTMock()
	svc := newTestService(t, pool, ytMock)

	_, _, err := svc.GetChannelUploads(ctx, userID, uuid.Must(uuid.NewV7()), nil, 10, "")
	if err == nil {
		t.Fatal("expected error for non-existent channel, got nil")
	}
}

func TestService_GetChannelDetail(t *testing.T) {
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

			var channelID uuid.UUID
			if tt.seed {
				ch := seedSubscription(t, pool, ytMock, userID)
				channelID = ch.ID
			} else {
				channelID = uuid.Must(uuid.NewV7())
			}

			svc := newTestService(t, pool, ytMock)
			detail, err := svc.GetChannelDetail(ctx, channelID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if detail.DisplayName != "Test Channel" {
				t.Fatalf("expected display name %q, got %q", "Test Channel", detail.DisplayName)
			}
			if detail.ChannelID != channelID {
				t.Fatalf("expected channel ID %s, got %s", channelID, detail.ChannelID)
			}
		})
	}
}

func TestService_GetChannelFeeds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		seedVC    bool
		wantCount int
	}{
		{name: "empty", seedVC: false, wantCount: 0},
		{name: "with_valuable_channel", seedVC: true, wantCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUser(t, pool)
			ytMock := defaultYTMock()

			if tt.seedVC {
				// チャンネルを登録してからValuableChannelとして保存
				ch := seedSubscription(t, pool, ytMock, userID)
				vc, err := NewValuableChannel(ch.ID, "education", "Great content")
				if err != nil {
					t.Fatalf("setup: NewValuableChannel failed: %v", err)
				}
				if _, err := NewValuableChannelRepository(sqlc.New(pool)).Save(ctx, vc); err != nil {
					t.Fatalf("setup: Save ValuableChannel failed: %v", err)
				}
			}

			svc := newTestService(t, pool, ytMock)
			feeds, err := svc.GetChannelFeeds(ctx)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(feeds) != tt.wantCount {
				t.Fatalf("expected %d feeds, got %d", tt.wantCount, len(feeds))
			}
		})
	}
}
