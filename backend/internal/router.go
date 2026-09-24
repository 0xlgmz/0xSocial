package internal

import (
	"net/http"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/httpapi"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/middleware0x"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	csrf := http.NewCrossOriginProtection()
	if err := csrf.AddTrustedOrigin("http://localhost:5173"); err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(csrf.Handler)

	// Routing GET requests
	r.Route("/api", func(r chi.Router) {
		r.HandleFunc("GET /health", httpapi.Health)

		// Routing Authentication related
		r.Route("/auth", func(r chi.Router) {
			r.HandleFunc("POST /logout", httpapi.Logout(pool))
			r.HandleFunc("POST /register", httpapi.Register(pool))
			r.HandleFunc("POST /login", httpapi.Login(pool))

			r.Group(func(r chi.Router) {
				r.Use(middleware0x.RequireAuth(pool))
				r.HandleFunc("GET /me", httpapi.CurrentSession)
			})
		})
	})
	return r
}
