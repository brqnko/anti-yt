package middleware_d

import (
	"context"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ScreenTimeMiddleware(db *pgxpool.Pool) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)
	userRepo := user.NewUserRepository(q)
	historyQS := history.NewHistoryQueryService(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			userID, err := hutil.UserIDFromContext(ctx)
			if err != nil {
				return f(ctx, w, r, request)
			}

			now := time.Now().UTC()

			rangeSet, err := userRepo.FindScreenTimeRanges(ctx, userID)
			if err != nil {
				hutil.LogError(ctx, err)
				return writeErrorJSON(w, http.StatusInternalServerError, "internal server error", "internal server error")
			}
			if nextStart := rangeSet.BlockedUntil(now); nextStart != nil {
				return writeForbiddenJSON(w, "outside_allowed_time_range", nextStart.Format(time.RFC3339))
			}

			watchStats, err := historyQS.FindTotalWatchSeconds(ctx, userID)
			if err != nil {
				hutil.LogError(ctx, err)
				return writeErrorJSON(w, http.StatusInternalServerError, "internal server error", "internal server error")
			}

			if watchStats.DailyLimitSeconds >= 24*60*60 {
				return f(ctx, w, r, request)
			}

			if watchStats.TodayWatchTotal >= watchStats.DailyLimitSeconds {
				tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
				return writeForbiddenJSON(w, "screen_time_limit_exceeded", tomorrow.Format(time.RFC3339))
			}

			return f(ctx, w, r, request)
		}
	}
}
