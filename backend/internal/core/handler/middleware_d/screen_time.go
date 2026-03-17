package middleware_d

import (
	"context"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ScreenTimeMiddleware(db *pgxpool.Pool) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			userID, err := util.UserIDFromContext(ctx)
			if err != nil {
				return f(ctx, w, r, request)
			}

			now := time.Now().UTC()

			ranges, err := q.GetUserScreenTimeRanges(ctx, userID)
			if err != nil {
				util.LogError(ctx, err)
				return writeErrorJSON(w, http.StatusInternalServerError, "internal server error", "internal server error")
			}
			if nextStart := blockedUntil(now, ranges); nextStart != nil {
				return writeForbiddenJSON(w, "outside_allowed_time_range", "outside allowed time range", "current time is outside the allowed screen time range", nextStart)
			}

			watchStats, err := q.GetTotalWatchSeconds(ctx, userID)
			if err != nil {
				util.LogError(ctx, err)
				return writeErrorJSON(w, http.StatusInternalServerError, "internal server error", "internal server error")
			}

			if watchStats.DailyLimitSeconds >= 24*60*60 {
				return f(ctx, w, r, request)
			}

			if watchStats.TodayWatchTotal >= watchStats.DailyLimitSeconds {
				tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
				return writeForbiddenJSON(w, "screen_time_limit_exceeded", "screen time limit exceeded", "daily screen time limit has been reached", &tomorrow)
			}

			return f(ctx, w, r, request)
		}
	}
}

func blockedUntil(now time.Time, ranges []sqlc.GetUserScreenTimeRangesRow) *time.Time {
	if len(ranges) == 0 {
		return nil
	}

	nowSeconds := database_d.Seconds(now.Hour()*3600 + now.Minute()*60 + now.Second())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	for _, r := range ranges {
		if nowSeconds >= r.ScreenTimeRangeStart && nowSeconds < r.ScreenTimeRangeEnd {
			return nil
		}
	}

	var best time.Time
	found := false
	for _, r := range ranges {
		if r.ScreenTimeRangeStart > nowSeconds {
			t := today.Add(time.Duration(r.ScreenTimeRangeStart) * time.Second)
			if !found || t.Before(best) {
				best = t
				found = true
			}
		}
	}
	if found {
		return &best
	}

	tomorrow := today.AddDate(0, 0, 1)
	for _, r := range ranges {
		t := tomorrow.Add(time.Duration(r.ScreenTimeRangeStart) * time.Second)
		if !found || t.Before(best) {
			best = t
			found = true
		}
	}
	if found {
		return &best
	}
	return nil
}
