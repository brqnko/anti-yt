package middleware_d

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
)

func TestScreenTimeMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		screenLimit   *int
		wantCalled    bool
		wantForbidden bool
	}{
		{
			name:       "unlimited_user",
			wantCalled: true,
		},
		{
			name:        "within_limit",
			screenLimit: intPtr(3600),
			wantCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool := testutil.NewTestDB(t)
			userID := testutil.SeedUserWithScreenLimit(t, pool, tt.screenLimit)

			called := false
			inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
				called = true
				return nil, nil
			}

			mwFactory := ScreenTimeMiddleware(pool)
			mw := mwFactory(inner, "test")

			ctx := hutil.WithUserID(context.Background(), userID)
			r := httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()

			_, _ = mw(ctx, w, r, nil)

			if called != tt.wantCalled {
				t.Fatalf("expected handler called=%v, got %v", tt.wantCalled, called)
			}
			if tt.wantForbidden && w.Code != http.StatusForbidden {
				t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
			}
		})
	}
}

func TestScreenTimeMiddleware_NoUserID(t *testing.T) {
	t.Parallel()

	pool := testutil.NewTestDB(t)

	called := false
	inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}

	mwFactory := ScreenTimeMiddleware(pool)
	mw := mwFactory(inner, "test")

	// userID なしのコンテキスト → スキップしてハンドラ呼び出し
	r := httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil)
	w := httptest.NewRecorder()

	_, _ = mw(context.Background(), w, r, nil)

	if !called {
		t.Fatal("handler should have been called when no userID in context")
	}
}

func intPtr(v int) *int { return &v }
