package middleware_d

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/jackc/pgx/v5"
)

func TestDomainErrorMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantTitle  string
	}{
		{
			name:       "no_error",
			err:        nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not_found",
			err:        pgx.ErrNoRows,
			wantStatus: http.StatusNotFound,
			wantTitle:  "Not Found",
		},
		{
			name:       "domain_error",
			err:        core.NewDomainError("INVALID_INPUT", "input is invalid"),
			wantStatus: http.StatusBadRequest,
			wantTitle:  "INVALID_INPUT",
		},
		{
			name:       "unexpected_error",
			err:        errors.New("something broke"),
			wantStatus: http.StatusInternalServerError,
			wantTitle:  "Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			var inner func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error)
			if tt.err != nil {
				inner = stubHandlerWithError(&called, tt.err)
			} else {
				inner = func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
					called = true
					w.WriteHeader(http.StatusOK)
					return "ok", nil
				}
			}

			mw := DomainErrorMiddleware(inner, "test")
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			_, _ = mw(context.Background(), w, r, nil)

			if !called {
				t.Fatal("handler should have been called")
			}
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if tt.wantTitle != "" {
				var body struct {
					Title string `json:"title"`
				}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if body.Title != tt.wantTitle {
					t.Fatalf("expected title %q, got %q", tt.wantTitle, body.Title)
				}
			}
		})
	}
}
