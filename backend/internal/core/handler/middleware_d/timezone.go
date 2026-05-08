package middleware_d

import (
	"context"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

func TimezoneMiddleware(
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		if _, ok := ignoreOperationIDs[operationID]; ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			tz := r.Header.Get("X-Timezone")
			if tz == "" {
				return writeErrorJSON(w, http.StatusBadRequest, "bad_request", "X-Timezone header is required")
			}
			loc, err := time.LoadLocation(tz)
			if err != nil {
				return writeErrorJSON(w, http.StatusBadRequest, "bad_request", "invalid timezone")
			}
			ctx = hutil.WithTimezone(ctx, loc)
			return f(ctx, w, r.WithContext(ctx), request)
		}
	}
}
