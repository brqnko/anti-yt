package middleware_d

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core"
	"github.com/brqnko/anti-yt/backend/internal/core/report"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5"
)

func DomainErrorMiddleware(reportService *report.Service) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
			response, err := f(ctx, w, r, request)
			if err == nil {
				return response, nil
			}

			if errors.Is(err, pgx.ErrNoRows) {
				return writeErrorJSON(w, http.StatusNotFound, "Not Found", "resource not found")
			}

			var domainErr *core.DomainError
			if errors.As(err, &domainErr) {
				return writeErrorJSON(w, http.StatusBadRequest, domainErr.Code(), domainErr.Error())
			}

			util.LoggerFromContext(ctx).ErrorContext(ctx, "internal server error", slog.Any("error", err))
			reportService.ReportError(ctx, operationID, err)
			return writeErrorJSON(w, http.StatusInternalServerError, "Internal Server Error", "an unexpected error has occurred")
		}
	}
}
