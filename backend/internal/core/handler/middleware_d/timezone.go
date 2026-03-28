package middleware_d

import (
	"context"
	"net/http"
	"time"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
)

var timezoneIgnoreOperations = map[string]struct{}{
	"GetAuthGoogle":               {},
	"GetAuthGoogleCallback":       {},
	"PostAuthLogout":              {},
	"PostAuthRefresh":             {},
	"GetHealth":                   {},
	"GetAuthOauthYoutube":         {},
	"GetAuthOauthYoutubeCallback": {},
}

func TimezoneMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	if _, ok := timezoneIgnoreOperations[operationID]; ok {
		return f
	}

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		tz := r.Header.Get("X-Timezone")
		if tz == "" {
			return writeErrorJSON(w, http.StatusBadRequest, "bad request", "X-Timezone header is required")
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return writeErrorJSON(w, http.StatusBadRequest, "bad request", "invalid timezone")
		}
		newCtx := hutil.WithTimezone(ctx, loc)
		return f(newCtx, w, r.WithContext(newCtx), request)
	}
}
