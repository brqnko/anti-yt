package middleware_d

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func TestResponseCookieMiddleware(t *testing.T) {
	t.Parallel()

	inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		hutil.AddResponseCookie(ctx, "token=abc; Path=/; HttpOnly")
		hutil.AddResponseCookie(ctx, "session=xyz; Path=/")
		return nil, nil
	}

	mw := ResponseCookieMiddleware(inner, "test")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	_, _ = mw(context.Background(), w, r, nil)

	cookies := w.Header().Values("Set-Cookie")
	if len(cookies) != 2 {
		t.Fatalf("expected 2 Set-Cookie headers, got %d", len(cookies))
	}
	if cookies[0] != "token=abc; Path=/; HttpOnly" {
		t.Fatalf("unexpected first cookie: %q", cookies[0])
	}
	if cookies[1] != "session=xyz; Path=/" {
		t.Fatalf("unexpected second cookie: %q", cookies[1])
	}
}

func TestResponseCookieMiddleware_NoCookies(t *testing.T) {
	t.Parallel()

	called := false
	inner := stubHandler(&called)

	mw := ResponseCookieMiddleware(inner, "test")
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	_, _ = mw(context.Background(), w, r, nil)

	if !called {
		t.Fatal("handler should have been called")
	}
	cookies := w.Header().Values("Set-Cookie")
	if len(cookies) != 0 {
		t.Fatalf("expected no Set-Cookie headers, got %d", len(cookies))
	}
}
