package internal

import (
	"net/http"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/httpapi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Routing GET requests
	r.Route("/api", func(r chi.Router) {
		routerGetters(r)
	})
	return r
}

func routerGetters(r chi.Router) {
	r.HandleFunc("GET /health", httpapi.Health)
}
