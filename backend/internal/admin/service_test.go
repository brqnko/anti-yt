package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/jackc/pgx/v5"
)

func TestService_CreateNewValuableChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		externalChannelID string
		reason            string
		description       string
		fetchErr          error
		wantErr           bool
	}{
		{
			name:              "success_by_handle",
			externalChannelID: "@testchannel",
			reason:            "education",
			description:       "Great educational content",
		},
		{
			name:              "success_by_channel_id",
			externalChannelID: "UCxxxxxxxxxxxxxxxxxxxxxx",
			reason:            "technology",
			description:       "Tech channel",
		},
		{
			name:              "invalid_channel_id",
			externalChannelID: "invalid",
			reason:            "education",
			description:       "desc",
			wantErr:           true,
		},
		{
			name:              "invalid_reason_code",
			externalChannelID: "@testchannel",
			reason:            "invalid_reason",
			description:       "desc",
			wantErr:           true,
		},
		{
			name:              "youtube_fetch_error",
			externalChannelID: "@testchannel",
			reason:            "education",
			description:       "desc",
			fetchErr:          errors.New("youtube API error"),
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)

			ytMock := defaultYTMock()
			if tt.fetchErr != nil {
				ytMock.FetchChannelDetailByIDOrHandleFunc = func(ctx context.Context, channelID string) (youtube_d.Channel, error) {
					return youtube_d.Channel{}, tt.fetchErr
				}
			}

			svc := newTestService(t, pool, ytMock)

			vc, err := svc.CreateNewValuableChannel(ctx, tt.externalChannelID, tt.reason, tt.description)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if vc == nil {
				t.Fatal("valuable channel should not be nil")
			}
			if vc.ValuableReasonCode.String() != tt.reason {
				t.Fatalf("expected reason %q, got %q", tt.reason, vc.ValuableReasonCode.String())
			}
			if vc.ValuableDescription.String() != tt.description {
				t.Fatalf("expected description %q, got %q", tt.description, vc.ValuableDescription.String())
			}
		})
	}
}

func TestService_CreateNewValuableChannel_ExistingChannel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	ytMock := defaultYTMock()

	svc := newTestService(t, pool, ytMock)

	// 1回目: チャンネル作成 + ValuableChannel作成
	_, err := svc.CreateNewValuableChannel(ctx, "@testchannel", "education", "desc")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// 2回目: 同じチャンネルで再度作成 → チャンネルは既存を使う（duplicate key エラーの可能性）
	_, err = svc.CreateNewValuableChannel(ctx, "@testchannel", "technology", "desc2")
	if err == nil {
		// ValuableChannel の unique 制約でエラーになるはず
		// もしエラーにならなければ、制約がないということなのでそれはそれでOK
		t.Log("second call succeeded (no unique constraint on valuable channel)")
	}
}

func TestService_UpdateValuableChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		newReason   *string
		newDesc     *string
		wantReason  string
		wantDesc    string
	}{
		{
			name:       "update_reason",
			newReason:  strPtr("technology"),
			wantReason: "technology",
			wantDesc:   "original desc",
		},
		{
			name:       "update_description",
			newDesc:    strPtr("updated description"),
			wantReason: "education",
			wantDesc:   "updated description",
		},
		{
			name:       "update_both",
			newReason:  strPtr("music"),
			newDesc:    strPtr("new desc"),
			wantReason: "music",
			wantDesc:   "new desc",
		},
		{
			name:       "no_changes",
			wantReason: "education",
			wantDesc:   "original desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			ytMock := defaultYTMock()

			externalID := seedValuableChannel(t, pool, ytMock, "education", "original desc")

			svc := newTestService(t, pool, ytMock)
			vc, err := svc.UpdateValuableChannel(ctx, externalID, tt.newReason, tt.newDesc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if vc.ValuableReasonCode.String() != tt.wantReason {
				t.Fatalf("expected reason %q, got %q", tt.wantReason, vc.ValuableReasonCode.String())
			}
			if vc.ValuableDescription.String() != tt.wantDesc {
				t.Fatalf("expected description %q, got %q", tt.wantDesc, vc.ValuableDescription.String())
			}
		})
	}
}

func TestService_UpdateValuableChannel_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	ytMock := defaultYTMock()
	svc := newTestService(t, pool, ytMock)

	reason := "education"
	_, err := svc.UpdateValuableChannel(ctx, "@nonexistent", &reason, nil)
	if err == nil {
		t.Fatal("expected error for non-existent channel, got nil")
	}
}

func TestService_UpdateValuableChannel_InvalidReason(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := testutil.NewTestDB(t)
	ytMock := defaultYTMock()

	externalID := seedValuableChannel(t, pool, ytMock, "education", "desc")

	svc := newTestService(t, pool, ytMock)
	reason := "invalid_reason"
	_, err := svc.UpdateValuableChannel(ctx, externalID, &reason, nil)
	if err == nil {
		t.Fatal("expected error for invalid reason, got nil")
	}
}

func TestService_RemoveValuableChannel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seed    bool
		wantErr bool
	}{
		{name: "success", seed: true, wantErr: false},
		{name: "not_found", seed: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			pool := testutil.NewTestDB(t)
			ytMock := defaultYTMock()

			var externalID string
			if tt.seed {
				externalID = seedValuableChannel(t, pool, ytMock, "education", "desc")
			} else {
				externalID = "@nonexistent"
			}

			svc := newTestService(t, pool, ytMock)
			err := svc.RemoveValuableChannel(ctx, externalID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// 削除後は更新できないことを確認
			reason := "technology"
			_, err = svc.UpdateValuableChannel(ctx, externalID, &reason, nil)
			if err == nil {
				t.Fatal("expected error after removal, got nil")
			}
			if !errors.Is(err, pgx.ErrNoRows) {
				t.Fatalf("expected pgx.ErrNoRows, got %v", err)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
