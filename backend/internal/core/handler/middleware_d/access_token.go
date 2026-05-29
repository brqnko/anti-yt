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

// access_tokenクッキーのJWTを解析し、user_idをcontextに付与する。
//
// ignoreOperationIDs: 認証オプショナル扱い。access_tokenクッキーが無く refresh_tokenも無い場合は匿名で通過。
// クッキーが存在して検証失敗、またはaccess_tokenが無く refresh_tokenが残っている場合は401を返す
// (フロントの自動リフレッシュ機構を起動させるため)。
//
// publicOperationIDs: ブラウザが直接遷移するリダイレクト型の認証フロー(Googleログイン開始/コールバック等)。
// access_tokenの有無に関わらず常に匿名で通過。refresh_tokenが残っていてもJSON 401を返さない。
//
// bypassOperationIDs: 検証失敗でも常に通過。register tokenを受け取るエンドポイント用。
func AccessTokenMiddleware(
	jwtService jwt_d.Service,
	jtiBlacklistRepo database_d.JtiBlacklistRepository,
	ignoreOperationIDs map[string]struct{},
	publicOperationIDs map[string]struct{},
	bypassOperationIDs map[string]struct{},
) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		_, optional := ignoreOperationIDs[operationID]
		_, public := publicOperationIDs[operationID]
		_, bypass := bypassOperationIDs[operationID]

		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
			cookie, err := r.Cookie("access_token")
			if err != nil {
				if bypass || public {
					// public: ブラウザが直接遷移するリダイレクト型の認証フロー
					// (Googleログイン開始/コールバック等)。access_tokenが無くても
					// JSON 401を返さず匿名で通過させる。
					return f(ctx, w, r, request)
				}
				// access_tokenが切れてブラウザに削除されたが refresh_tokenが残っている場合、
				// 401を返してフロントの自動リフレッシュ機構を起動させる
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
				if bypass || public {
					return f(ctx, w, r.WithContext(ctx), request)
				}
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
