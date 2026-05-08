package middleware_d

import (
	"context"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

func AuthTokensMiddleware(
	requireOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		if _, ok := requireOperationIDs[operationID]; !ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			if cookie, err := r.Cookie("access_token"); err == nil {
				ctx = hutil.WithAccessToken(ctx, cookie.Value)
			}
			if cookie, err := r.Cookie("refresh_token"); err == nil {
				ctx = hutil.WithRefreshToken(ctx, cookie.Value)
			}
			return f(ctx, w, r.WithContext(ctx), request)
		}
	}
}
