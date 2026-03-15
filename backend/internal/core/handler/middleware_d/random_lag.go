package middleware_d

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
)

func RandomLagMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		lag := 300 + rand.IntN(201) // 300~500ms
		time.Sleep(time.Duration(lag) * time.Millisecond)
		return f(ctx, w, r, request)
	}
}
