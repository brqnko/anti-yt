package middleware

import (
	"net/http"
	"strings"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/labstack/echo/v4"
)

var (
	// 一応
	csrfExcludedPathPrefixes = []string{
		"/api/v1/auth/google",
		"/api/v1/auth/google/callback",
	}
)

func CsrfMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(c echo.Context, request interface{}) (response interface{}, err error) {
		method := c.Request().Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions || method == http.MethodTrace {
			return f(c, request)
		}
		for _, prefix := range csrfExcludedPathPrefixes {
			if strings.HasPrefix(c.Path(), prefix) {
				return f(c, request)
			}
		}

		cookie, err := c.Cookie("csrf_token")
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "bad request")
		}
		header := c.Request().Header.Get("x-csrf-token")
		if header == "" || cookie.Value == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "bad request")
		}
		if header != cookie.Value {
			return nil, echo.NewHTTPError(http.StatusBadRequest, "bad request")
		}

		return f(c, request)
	}
}
