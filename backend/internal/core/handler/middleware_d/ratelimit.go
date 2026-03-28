package middleware_d

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/brqnko/anti-yt/backend/internal/core/database_d/sqlc"
	"github.com/brqnko/anti-yt/backend/internal/core/handler/hutil"
	v1 "github.com/brqnko/anti-yt/backend/internal/core/handler/v1"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type quotaMapKey struct {
	method    string
	pathRegex *regexp.Regexp
	quota     int
}

func buildQuotaMap(keys []quotaMapKey, r *chi.Mux) map[string]int {
	mp := make(map[string]int)
	err := chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "" {
			panic("route is empty")
		}
		for _, k := range keys {
			if method != k.method {
				continue
			}
			if k.pathRegex.MatchString(route) {
				mp[fmt.Sprintf("%s:%s", method, route)] = k.quota
				break
			}
		}

		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("chi.Walk failed: %v", err))
	}
	if len(mp) == 0 {
		panic("quotaMap is empty: no routes matched")
	}
	return mp
}

// ctxにuserIDがあるなら、そのユーザーのレートリミットを確認&更新する
func UserRatelimitMiddleware(mux *chi.Mux, db *pgxpool.Pool) func(v1.StrictHandlerFunc, string) v1.StrictHandlerFunc {
	q := sqlc.New(db)

	userQuotaLimit := 2000
	keys := []quotaMapKey{
		{
			method:    http.MethodPost,
			pathRegex: regexp.MustCompile(`/api/v1/channels/subscribe$`),
			quota:     3,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/channels/\{[^/]+\}/videos$`),
			quota:     2,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/channels/\{[^/]+\}$`),
			quota:     1,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/feed/channels$`),
			quota:     3,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/feed$`),
			quota:     2,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/search$`),
			quota:     100,
		},
		{
			method:    http.MethodPost,
			pathRegex: regexp.MustCompile(`/api/v1/playlists/\{[^/]+\}/videos$`),
			quota:     100,
		},
		{
			method:    http.MethodGet,
			pathRegex: regexp.MustCompile(`/api/v1/auth/oauth/youtube/callback$`),
			quota:     500,
		},
	}
	var quotaMap map[string]int
	var once sync.Once

	return func(f v1.StrictHandlerFunc, operationID string) v1.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
			once.Do(func() {
				quotaMap = buildQuotaMap(keys, mux)
			})

			userID, err := hutil.UserIDFromContext(ctx)
			if err != nil {
				return f(ctx, w, r, request)
			}

			if r.Pattern == "" {
				return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "An unexpected error occurred.")
			}

			quota := 1 // 見つからない場合は1クオータ
			if found, ok := quotaMap[fmt.Sprintf("%s:%s", r.Method, r.Pattern)]; ok {
				quota = found
			}
			row, err := q.UpsertRatelimit(ctx, sqlc.UpsertRatelimitParams{
				UserID: userID,
				Quota:  quota,
			})
			if err != nil {
				return writeErrorJSON(w, http.StatusInternalServerError, "internal_server_error", "An error occurred while checking the rate limit.")
			}
			// リクエスト前の消費量が既に上限以上なら拒否
			if row.ConsumedQuota-quota >= userQuotaLimit {
				return writeErrorJSON(w, http.StatusTooManyRequests, "too_many_requests", "Rate limit exceeded. Please try again later.")
			}

			return f(ctx, w, r, request)
		}
	}
}
