//go:generate moq -out mock_jwt_service_test.go -pkg middleware_d_test ../../jwt_d Service
//go:generate moq -out mock_jti_blacklist_repository_test.go -pkg middleware_d_test ../../database_d JtiBlacklistRepository

package middleware_d

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

// access_tokenヘッダのJWTを解析し、user_idをcontextに付与する
// 引数のmapにマッチしたリクエストは解析を無視される
func AccessTokenMiddleware(
	jwtService jwt_d.Service,
	jtiBlacklistRepo database_d.JtiBlacklistRepository,
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		if _, ok := ignoreOperationIDs[operationID]; ok {
			// 認証をオプショナルにする: JWTがあればuser_idをcontextに付与するが、なくても続行する
			return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
				cookie, err := r.Cookie("access_token")
				if err != nil {
					return f(ctx, w, r, request)
				}
				userID, jti, _, err := jwtService.VerifyUserAccessToken(cookie.Value)
				if err != nil {
					return f(ctx, w, r, request)
				}
				blacklisted, err := jtiBlacklistRepo.IsJtiExist(ctx, jti)
				if err != nil || blacklisted {
					return f(ctx, w, r, request)
				}
				newCtx := hutil.WithUserID(ctx, userID)
				return f(newCtx, w, r.WithContext(newCtx), request)
			}
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

			blacklisted, err := jtiBlacklistRepo.IsJtiExist(ctx, jti)
			if err != nil {
				util.LoggerFromContext(ctx).ErrorContext(ctx, "failed to look up jti blacklist", slog.Any("error", err))
				return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "An error occurred while checking the access token.")
			}
			if blacklisted {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}

			newCtx := hutil.WithUserID(ctx, userID)
			return f(newCtx, w, r.WithContext(newCtx), request)
		}
	}
}
