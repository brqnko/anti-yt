package middleware_d

import (
	"context"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

// ctxにuserIDがあるなら、そのユーザーのレートリミットを確認&更新する
// operationIDQuotasに登録されていないoperationIDは1quota扱い。
// userQuotaLimitは1ウィンドウあたりに使えるquotaの上限。
func UserRatelimitMiddleware(
	repo database_d.RatelimitRepository,
	userQuotaLimit int,
	operationIDQuotas map[string]int,
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
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
			consumed, err := repo.Consume(ctx, userID, quota)
			if err != nil {
				return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "An error occurred while checking the rate limit.")
			}
			// リクエスト前の消費量が既に上限以上なら拒否
			if consumed-quota >= userQuotaLimit {
				return writeErrorJSON(w, http.StatusTooManyRequests, "too_many_requests", "Rate limit exceeded. Please try again later.")
			}

			return f(ctx, w, r, request)
		}
	}
}
