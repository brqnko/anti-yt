package middleware_d

import (
	"context"
	"net/http"
	"strings"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

var (
	// 一応
	csrfExcludedPathPrefixes = []string{
		"/api/v1/auth/google",
		"/api/v1/auth/google/callback",
	}
)

func CsrfMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		method := r.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodTrace {
			return f(ctx, w, r, request)
		}
		for _, prefix := range csrfExcludedPathPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return f(ctx, w, r, request)
			}
		}

		cookie, err := r.Cookie("csrf_token")
		if err != nil {
			return writeErrorJSON(w, http.StatusBadRequest, "bad request", "bad request")
		}
		header := r.Header.Get("x-csrf-token")
		if header == "" || cookie.Value == "" {
			return writeErrorJSON(w, http.StatusBadRequest, "bad request", "bad request")
		}
		if header != cookie.Value {
			return writeErrorJSON(w, http.StatusBadRequest, "bad request", "bad request")
		}

		return f(ctx, w, r, request)
	}
}
