package middleware_d

import (
	"context"
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/google/uuid"
)

// contextにrequest_id(uuid v7)を付与する
func RequestIDMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (_ interface{}, err error) {
		requestID, err := uuid.NewV7()
		if err != nil {
			return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "failed to generate request id")
		}

		newCtx := hutil.WithRequestID(ctx, requestID)
		newCtx = hutil.WithRequestPath(newCtx, r.URL.Path)
		return f(newCtx, w, r.WithContext(newCtx), request)
	}
}
