package middleware

import (
	"context"
	"strings"

	"github.com/brqnko/anti-yt/backend/internal/core/handler"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/labstack/echo/v4"
)

var (
	requestTokensRequiredPathPrefixes = []string{
		"/api/v1/auth/logout",
		"/api/v1/auth/refresh",
	}
)

func AuthTokensMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(c echo.Context, request interface{}) (response interface{}, err error) {
		required := false
		for _, prefix := range requestTokensRequiredPathPrefixes {
			if strings.HasPrefix(c.Path(), prefix) {
				required = true
				break
			}
		}
		if !required {
			return f(c, request)
		}

		req := c.Request()
		ctx := req.Context()
		if cookie, err := req.Cookie("access_token"); err == nil {
			ctx = context.WithValue(ctx, handler.AccessTokenKey{}, cookie.Value)
		}
		if cookie, err := req.Cookie("refresh_token"); err == nil {
			ctx = context.WithValue(ctx, handler.RefreshTokenKey{}, cookie.Value)
		}
		c.SetRequest(req.WithContext(ctx))
		return f(c, request)
	}
}
