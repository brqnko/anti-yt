package middleware_d

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCsrfMiddleware(t *testing.T) {
	t.Parallel()

	handlerCalled := false
	inner := stubHandler(&handlerCalled)

	tests := []struct {
		name       string
		method     string
		path       string
		cookie     string
		header     string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "get_request_skipped",
			method:     http.MethodGet,
			path:       "/api/v1/users",
			wantStatus: 0,
			wantCalled: true,
		},
		{
			name:       "excluded_path_google_auth",
			method:     http.MethodPost,
			path:       "/api/v1/auth/google/callback",
			wantStatus: 0,
			wantCalled: true,
		},
		{
			name:       "valid_csrf",
			method:     http.MethodPost,
			path:       "/api/v1/users",
			cookie:     "valid-token",
			header:     "valid-token",
			wantStatus: 0,
			wantCalled: true,
		},
		{
			name:       "missing_cookie",
			method:     http.MethodPost,
			path:       "/api/v1/users",
			header:     "token",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "missing_header",
			method:     http.MethodPost,
			path:       "/api/v1/users",
			cookie:     "token",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
		{
			name:       "mismatch",
			method:     http.MethodPost,
			path:       "/api/v1/users",
			cookie:     "token-a",
			header:     "token-b",
			wantStatus: http.StatusBadRequest,
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handlerCalled = false

			mw := CsrfMiddleware(inner, "test")

			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			if tt.cookie != "" {
				r.AddCookie(&http.Cookie{Name: "csrf_token", Value: tt.cookie})
			}
			if tt.header != "" {
				r.Header.Set("x-csrf-token", tt.header)
			}

			_, _ = mw(context.Background(), w, r, nil)

			if tt.wantCalled != handlerCalled {
				t.Fatalf("expected handler called=%v, got %v", tt.wantCalled, handlerCalled)
			}
			if tt.wantStatus != 0 && w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
