package middleware

import (
	"net/http"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func RequestIDMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx echo.Context, request interface{}) (response interface{}, err error) {
		requestID, err := uuid.NewV7()
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
		}
		req := ctx.Request()
		newCtx := util.WithRequestID(req.Context(), requestID)
		ctx.SetRequest(req.WithContext(newCtx))
		return f(ctx, request)
	}
}
