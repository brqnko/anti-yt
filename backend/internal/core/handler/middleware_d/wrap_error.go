package middleware_d

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/discord_d"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

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

func WrapErrorMiddleware(discordClient discord_d.Client) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
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

			reportCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			go func() {
				defer cancel()
				if sendErr := discordClient.SendWebhookMessage(reportCtx, fmt.Sprintf("[ERROR] `%s`\n```\n%v\n```", operationID, err)); sendErr != nil {
					util.LoggerFromContext(reportCtx).ErrorContext(reportCtx, "failed to send error report", slog.Any("error", sendErr))
				}
			}()

			return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "an unexpected error has occurred")
		}
	}
}
