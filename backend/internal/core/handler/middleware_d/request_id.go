package middleware_d

import (
	"context"
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
)

func RequestIDMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		requestID, err := uuid.NewV7()
		if err != nil {
			return writeErrorJSON(w, http.StatusInternalServerError, "internal server error", "internal server error")
		}
		newCtx := util.WithRequestID(ctx, requestID)
		newCtx = util.WithRequestPath(newCtx, r.URL.Path)
		return f(newCtx, w, r.WithContext(newCtx), request)
	}
}
