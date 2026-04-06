//go:generate moq -out mock_jwt_service_test.go -pkg middleware_d_test ../../jwt_d Service

package middleware_d

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// access_tokenヘッダのJWTを解析し、user_idをcontextに付与する
// 引数のmapにマッチしたリクエストは解析を無視される
func AccessTokenMiddleware(
	jwtService jwt_d.Service,
	db *pgxpool.Pool,
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		if _, ok := ignoreOperationIDs[operationID]; ok {
			return f
		}

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}
			userID, jti, _, err := jwtService.VerifyUserAccessToken(cookie.Value)
			if err != nil {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}

			expiresAt, err := q.FindBlacklistedJTI(ctx, jti)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}
			// ErrNoRows = ブラックリストに無い: アクセス許可
			// レコードが存在しても有効期限切れ: アクセス許可
			if errors.Is(err, pgx.ErrNoRows) || !time.Now().UTC().Before(expiresAt) {
				newCtx := hutil.WithUserID(ctx, userID)
				return f(newCtx, w, r.WithContext(newCtx), request)
			}

			// ブラックリストに存在し、まだ有効期限内: 拒否
			return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		}
	}
}
