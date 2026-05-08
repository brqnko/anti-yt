package middleware_d

import (
	"context"
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

func ResponseCookieMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		ctx = hutil.WithResponseCookies(ctx)
		resp, err := f(ctx, w, r.WithContext(ctx), request)
		for _, cookie := range hutil.ResponseCookiesFromContext(ctx) {
			w.Header().Add("Set-Cookie", cookie)
		}
		return resp, err
	}
}
