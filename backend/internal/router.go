package internal

import (
	"net/http"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/httpapi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Routing GET requests
	r.Route("/api", func(r chi.Router) {
		routerGetters(r, pool)

		// Routing Authentication related
		r.Route("/auth", func(r chi.Router) {
			// Routing POST requests
			routerPosters(r, pool)
		})
	})
	return r
}

func routerGetters(r chi.Router, pool *pgxpool.Pool) {
	r.HandleFunc("GET /health", httpapi.Health)
}
func routerPosters(r chi.Router, pool *pgxpool.Pool) {
	r.HandleFunc("POST /register", httpapi.Register(pool))
}
