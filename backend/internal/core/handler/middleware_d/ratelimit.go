package middleware_d

import (
	"context"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ctxにuserIDがあるなら、そのユーザーのレートリミットを確認&更新する
// operationIDQuotasに登録されていないoperationIDは1quota扱い。
// userQuotaLimitは1ユーザが1日あたりに使えるquotaの上限。
func UserRatelimitMiddleware(
	db *pgxpool.Pool,
	userQuotaLimit int,
	operationIDQuotas map[string]int,
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			userID, err := hutil.UserIDFromContext(ctx)
			if err != nil {
				return f(ctx, w, r, request)
			}

			quota := 1 // 見つからない場合は1クオータ
			if found, ok := operationIDQuotas[operationID]; ok {
				quota = found
			}
			row, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{
				UserID: userID,
				Quota:  quota,
			})
			if err != nil {
				return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "An error occurred while checking the rate limit.")
			}
			// リクエスト前の消費量が既に上限以上なら拒否
			if row.ConsumedQuota-quota >= userQuotaLimit {
				return writeErrorJSON(w, http.StatusTooManyRequests, "too_many_requests", "Rate limit exceeded. Please try again later.")
			}

			return f(ctx, w, r, request)
		}
	}
}
