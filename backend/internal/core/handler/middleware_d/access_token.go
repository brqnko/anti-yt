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

func AccessTokenMiddleware(
	jwtService jwt_d.Service,
	jtiBlacklistRepo database_d.JtiBlacklistRepository,
	ignoreOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		_, optional := ignoreOperationIDs[operationID]

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				if operationID == "PostAuthRefresh" {
					return f(ctx, w, r, request)
				}
				if _, rerr := r.Cookie("refresh_token"); rerr == nil {
					return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
				}
				if optional {
					return f(ctx, w, r, request)
				}
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

			ctx = hutil.WithUserID(ctx, userID)
			return f(ctx, w, r.WithContext(ctx), request)
		}
	}
}
