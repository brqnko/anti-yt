package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleAdminEndpoints(m *chi.Mux, db *pgxpool.Pool, adminAPIKey string) {
	m.Route("/admin/api/v1", func(r chi.Router) {
		r.Use(adminAPIKeyAuthMiddleware(adminAPIKey))
		// TODO: ルートを追加
	})
}

func adminAPIKeyAuthMiddleware(adminAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token != "Bearer "+adminAPIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
