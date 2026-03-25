package middleware_d

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func TestAuthTokensMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		method          string
		path            string
		accessToken     string
		refreshToken    string
		wantAccess      bool
		wantRefresh     bool
	}{
		{
			name:   "non_required_path_skipped",
			method: http.MethodGet,
			path:   "/api/v1/feed",
		},
		{
			name:         "logout_path_extracts_tokens",
			method:       http.MethodPost,
			path:         "/api/v1/auth/logout",
			accessToken:  "at-123",
			refreshToken: "rt-456",
			wantAccess:   true,
			wantRefresh:  true,
		},
		{
			name:         "refresh_path_extracts_tokens",
			method:       http.MethodPost,
			path:         "/api/v1/auth/refresh",
			accessToken:  "at-789",
			refreshToken: "rt-012",
			wantAccess:   true,
			wantRefresh:  true,
		},
		{
			name:         "user_creation_extracts_tokens",
			method:       http.MethodPost,
			path:         "/api/v1/users",
			accessToken:  "at-abc",
			refreshToken: "rt-def",
			wantAccess:   true,
			wantRefresh:  true,
		},
		{
			name:   "required_path_no_cookies",
			method: http.MethodPost,
			path:   "/api/v1/auth/logout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var gotCtx context.Context
			inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
				gotCtx = ctx
				return nil, nil
			}

			mw := AuthTokensMiddleware(inner, "test")
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			if tt.accessToken != "" {
				r.AddCookie(&http.Cookie{Name: "access_token", Value: tt.accessToken})
			}
			if tt.refreshToken != "" {
				r.AddCookie(&http.Cookie{Name: "refresh_token", Value: tt.refreshToken})
			}

			_, _ = mw(context.Background(), w, r, nil)

			if tt.wantAccess {
				at, ok := hutil.AccessTokenFromContext(gotCtx)
				if !ok {
					t.Fatal("expected access token in context")
				}
				if at != tt.accessToken {
					t.Fatalf("expected access token %q, got %q", tt.accessToken, at)
				}
			}
			if tt.wantRefresh {
				rt, ok := hutil.RefreshTokenFromContext(gotCtx)
				if !ok {
					t.Fatal("expected refresh token in context")
				}
				if rt != tt.refreshToken {
					t.Fatalf("expected refresh token %q, got %q", tt.refreshToken, rt)
				}
			}
		})
	}
}
