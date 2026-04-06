//go:generate moq -out mock_discord_service_test.go -pkg middleware_d_test ../../discord_d Service:DiscordServiceMock

package middleware_d

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/report"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

// writeErrorJSON はステータスコード・タイトル・詳細を {title, detail} 形式の
// JSONとして ResponseWriter に直接書き込む。middleware内でエラーレスポンスを
// 短絡的に返すときに使う。StrictHandlerFunc のシグネチャに合わせて (nil, nil)
// を返し、呼び出し側には「自分で書き込んだので上層は何もしないで」と伝える。
func writeErrorJSON(w http.ResponseWriter, statusCode int, title, detail string) (interface{}, error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(struct {
		Title  string `json:"title"`
		Detail string `json:"detail"`
	}{
		Title:  title,
		Detail: detail,
	})
	return nil, nil
}

func WrapErrorMiddleware(reportService *report.Service) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			response, err := f(ctx, w, r, request)
			if err == nil {
				return response, nil
			}

			var domainErr *core.DomainError
			if errors.As(err, &domainErr) {
				return writeErrorJSON(w, domainErr.StatusCode().Int(), domainErr.Code(), domainErr.Error())
			}

			util.LoggerFromContext(ctx).ErrorContext(ctx, "internal server error",
				slog.String("operation_id", operationID),
				slog.Any("error", err),
			)
			reportService.ReportError(ctx, operationID, err)

			return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "an unexpected error has occurred")
		}
	}
}
