package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/brqnko/anti-yt/backend/internal/core/database/sqlc"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/brqnko/anti-yt/backend/internal/core/jwt_d"
	"github.com/brqnko/anti-yt/backend/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	accessTokenIgnorePathPrefixes = []string{
		"/api/v1/auth/google",
		"/api/v1/auth/google/callback",
	}
)

func AccessTokenMiddleware(jwtService jwt_d.JWTService, db *pgxpool.Pool) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			for _, prefix := range accessTokenIgnorePathPrefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					return f(ctx, w, r, request)
				}
			}
			if r.URL.Path == "/api/v1/users/me" && r.Method == http.MethodPost {
				return f(ctx, w, r, request)
			}

			cookie, err := r.Cookie("access_token")
			if err != nil {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}
			userID, jti, _, err := jwtService.VerifyUserAccessToken(cookie.Value)
			if err != nil {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}
			expiresAt, err := q.IsJTIBlacklisted(ctx, jti)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			}

			if errors.Is(err, pgx.ErrNoRows) || !time.Now().Before(expiresAt) {
				newCtx := util.WithUserID(ctx, userID)
				return f(newCtx, w, r.WithContext(newCtx), request)
			}

			return writeErrorJSON(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		}
	}
}
