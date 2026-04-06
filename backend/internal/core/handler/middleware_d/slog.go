package middleware_d

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

// HTTPリクエスト情報,request_id, user_idの情報をもったslogをcontextに付与する
// uuid.UUID型の値はhandlerのReplaceAttrで自動的にbase64形式に変換される
func SlogMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		attrs := []any{
			slog.Group("http",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			),
		}

		if requestID, ok := hutil.RequestIDFromContext(ctx); ok {
			attrs = append(attrs, slog.Any("request_id", requestID))
		}
		if userID, err := hutil.UserIDFromContext(ctx); err == nil {
			attrs = append(attrs, slog.Any("user_id", userID))
		}

		logger := slog.Default().With(attrs...)
		newCtx := util.WithLogger(ctx, logger)
		return f(newCtx, w, r.WithContext(newCtx), request)
	}
}
