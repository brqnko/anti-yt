package middleware

import (
	"crypto/ed25519"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/claims"
	"github.com/brqnko/anti-yt/backend/internal/core/database/sqlc"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

var (
	accessTokenIgnorePathPrefixes = []string{
		"/api/v1/auth/google",
		"/api/v1/auth/google/callback",
	}
)

func AccessTokenMiddleware(jwtPublic ed25519.PublicKey, db *pgxpool.Pool) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx echo.Context, request interface{}) (response interface{}, err error) {
			for _, prefix := range accessTokenIgnorePathPrefixes {
				if strings.HasPrefix(ctx.Path(), prefix) {
					return f(ctx, request)
				}
			}

			cookie, err := ctx.Cookie("access_token")
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			userID, jti, _, err := claims.VerifyUserAccessTokenWithExpiry(jwtPublic, cookie.Value)
			if err != nil {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}
			req := ctx.Request()
			expiresAt, err := q.IsJTIBlacklisted(req.Context(), jti)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
			}

			if errors.Is(err, pgx.ErrNoRows) || !time.Now().Before(expiresAt) {
				newCtx := util.WithUserID(req.Context(), userID)
				ctx.SetRequest(req.WithContext(newCtx))
				return f(ctx, request)
			}

			return nil, echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
		}
	}
}
