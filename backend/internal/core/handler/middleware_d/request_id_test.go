package middleware_d

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/google/uuid"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Parallel()

	var gotCtx context.Context
	inner := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		gotCtx = ctx
		return nil, nil
	}

	mw := RequestIDMiddleware(inner, "test")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	_, _ = mw(context.Background(), w, r, nil)

	reqID, ok := hutil.RequestIDFromContext(gotCtx)
	if !ok {
		t.Fatal("expected request ID in context")
	}
	if reqID == uuid.Nil {
		t.Fatal("request ID should not be nil UUID")
	}

	path, ok := hutil.RequestPathFromContext(gotCtx)
	if !ok {
		t.Fatal("expected request path in context")
	}
	if path != "/api/v1/users" {
		t.Fatalf("expected path %q, got %q", "/api/v1/users", path)
	}
}
