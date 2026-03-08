package middleware

import (
	"context"
	"net/http"
	"strings"

	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
)

var (
	requestTokensRequiredPathPrefixes = []string{
		"/api/v1/auth/logout",
		"/api/v1/auth/refresh",
	}
)

func AuthTokensMiddleware(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
		required := false
		for _, prefix := range requestTokensRequiredPathPrefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				required = true
				break
			}
		}
		required = required || (r.URL.Path == "/api/v1/users/me" && r.Method == http.MethodPost)
		if !required {
			return f(ctx, w, r, request)
		}

		newCtx := ctx
		if cookie, err := r.Cookie("access_token"); err == nil {
			newCtx = util.WithAccessToken(newCtx, cookie.Value)
		}
		if cookie, err := r.Cookie("refresh_token"); err == nil {
			newCtx = util.WithRefreshToken(newCtx, cookie.Value)
		}
		return f(newCtx, w, r.WithContext(newCtx), request)
	}
}
