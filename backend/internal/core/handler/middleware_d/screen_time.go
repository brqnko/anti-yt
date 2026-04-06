package middleware_d

import (
	"context"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/history"
	"github.com/brqnko/anti-yt/backend/internal/user"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ScreenTimeMiddleware(
	db *pgxpool.Pool,
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		// プロフィール設定などは制限の対象外
		if _, ok := ignoreOperationIDs[operationID]; ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			userID, err := hutil.UserIDFromContext(ctx)
			if err != nil {
				return f(ctx, w, r, request)
			}

			loc := hutil.TimezoneFromContext(ctx)
			now := time.Now()

			// 許可時間帯はUTCで保存されているのでUTCで比較
			rangeSet, err := user.NewUserRepository(sqlc.New(db)).FindScreenTimeRanges(ctx, userID)
			if err != nil {
				return nil, err
			}
			if nextStart := rangeSet.BlockedUntil(now.UTC()); nextStart != nil {
				return nil, core.NewDomainError(
					"outside_allowed_time_range",
					nextStart.Format(time.RFC3339),
					core.StatusForbidden,
				)
			}

			// 日次視聴制限はユーザーのローカル日付で計算
			remainingSeconds, err := history.NewHistoryQueryService(db).FindTotalWatchSeconds(ctx, userID, loc)
			if err != nil {
				return nil, err
			}

			if remainingSeconds <= 0 {
				localNow := now.In(loc)
				tomorrow := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, loc)
				return nil, core.NewDomainError(
					"screen_time_limit_exceeded",
					tomorrow.Format(time.RFC3339),
					core.StatusForbidden,
				)
			}

			return f(ctx, w, r, request)
		}
	}
}
