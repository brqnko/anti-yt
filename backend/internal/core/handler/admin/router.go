package admin

import (
	"net/http"

	"github.com/brqnko/anti-yt/backend/internal/channel"
	"github.com/brqnko/anti-yt/backend/internal/core/database_d"
	"github.com/brqnko/anti-yt/backend/internal/core/youtube_d"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleAdminEndpoints(m *chi.Mux, db *pgxpool.Pool, youtubeClient youtube_d.Client, feedRepo database_d.FeedRepository, adminAPIKey string) {
	h := newHandler(channel.NewService(db, youtubeClient, feedRepo))

	m.Route("/api/admin", func(r chi.Router) {
		r.Use(adminAPIKeyAuthMiddleware(adminAPIKey))
		r.Post("/valuable", h.createValuableChannel)
		r.Patch("/valuable", h.updateValuableChannel)
		r.Delete("/valuable", h.removeValuableChannel)
	})
}

func adminAPIKeyAuthMiddleware(adminAPIKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer "+adminAPIKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
