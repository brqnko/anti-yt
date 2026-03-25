package middleware_d

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/brqnko/anti-yt/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestAccessTokenMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		cookie     string
		verifyErr  error
		wantStatus int
		wantCalled bool
		wantUserID bool
	}{
		{
			name:       "excluded_path_google_auth",
			method:     http.MethodGet,
			path:       "/api/v1/auth/google",
			wantCalled: true,
		},
		{
			name:       "excluded_path_refresh",
			method:     http.MethodPost,
			path:       "/api/v1/auth/refresh",
			wantCalled: true,
		},
		{
			name:       "excluded_user_creation",
			method:     http.MethodPost,
			path:       "/api/v1/users",
			wantCalled: true,
		},
		{
			name:       "missing_cookie",
			method:     http.MethodGet,
			path:       "/api/v1/feed",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "invalid_token",
			method:     http.MethodGet,
			path:       "/api/v1/feed",
			cookie:     "bad-token",
			verifyErr:  errors.New("invalid"),
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "valid_token",
			method:     http.MethodGet,
			path:       "/api/v1/feed",
			cookie:     "good-token",
			wantCalled: true,
			wantUserID: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool := testutil.NewTestDB(t)
			userID := uuid.Must(uuid.NewV7())
			jti := uuid.Must(uuid.NewV7())

			jwtMock := testutil.DefaultJWTMock()
			if tt.verifyErr != nil {
				jwtMock.VerifyUserAccessTokenFunc = func(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
					return uuid.Nil, uuid.Nil, time.Time{}, tt.verifyErr
				}
			} else {
				jwtMock.VerifyUserAccessTokenFunc = func(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
					return userID, jti, time.Now().Add(30 * time.Minute), nil
				}
			}

			var gotCtx context.Context
			called := false
			inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
				called = true
				gotCtx = ctx
				return nil, nil
			}

			mwFactory := AccessTokenMiddleware(jwtMock, pool)
			mw := mwFactory(inner, "test")

			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{Name: "access_token", Value: tt.cookie})
			}

			_, _ = mw(context.Background(), w, r, nil)

			if called != tt.wantCalled {
				t.Fatalf("expected handler called=%v, got %v", tt.wantCalled, called)
			}
			if tt.wantStatus != 0 && w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantUserID {
				gotUserID, err := hutil.UserIDFromContext(gotCtx)
				if err != nil {
					t.Fatalf("expected userID in context: %v", err)
				}
				if gotUserID != userID {
					t.Fatalf("expected userID %s, got %s", userID, gotUserID)
				}
			}
		})
	}
}

func TestAccessTokenMiddleware_BlacklistedJTI(t *testing.T) {
	t.Parallel()

	pool := testutil.NewTestDB(t)
	userID := uuid.Must(uuid.NewV7())
	jti := uuid.Must(uuid.NewV7())

	// JTI をブラックリストに追加
	_, err := pool.Exec(context.Background(),
		"INSERT INTO t_jti_blacklist (jti, expires_at) VALUES ($1, $2)",
		jti, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("setup: insert blacklist: %v", err)
	}

	jwtMock := testutil.DefaultJWTMock()
	jwtMock.VerifyUserAccessTokenFunc = func(token string) (uuid.UUID, uuid.UUID, time.Time, error) {
		return userID, jti, time.Now().Add(30 * time.Minute), nil
	}

	called := false
	inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}

	mwFactory := AccessTokenMiddleware(jwtMock, pool)
	mw := mwFactory(inner, "test")

	r := httptest.NewRequest(http.MethodGet, "/api/v1/feed", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "blacklisted-token"})
	w := httptest.NewRecorder()

	_, _ = mw(context.Background(), w, r, nil)

	if called {
		t.Fatal("handler should NOT have been called for blacklisted JTI")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
