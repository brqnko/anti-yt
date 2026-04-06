package middleware_d

import (
	"context"
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

// double-submit cookieパターンでCSRFを検証する。
// GET/HEAD/OPTIONS/TRACE等の安全メソッドは自動的にスキップされる。
// ignoreOperationIDsに含まれるoperationIDも検証対象外。
func CsrfMiddleware(
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		if _, ok := ignoreOperationIDs[operationID]; ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
				return f(ctx, w, r, request)
			}

			cookie, err := r.Cookie("csrf_token")
			if err != nil || cookie.Value == "" {
				return writeErrorJSON(w, http.StatusBadRequest, "csrf_cookie_missing", "csrf_token cookie is missing or empty")
			}
			header := r.Header.Get("x-csrf-token")
			if header == "" {
				return writeErrorJSON(w, http.StatusBadRequest, "csrf_header_missing", "x-csrf-token header is missing")
			}
			if header != cookie.Value {
				return writeErrorJSON(w, http.StatusBadRequest, "csrf_mismatch", "x-csrf-token header does not match csrf_token cookie")
			}

			return f(ctx, w, r, request)
		}
	}
}
