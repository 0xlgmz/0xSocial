package internal

import (
	"net/http"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/httpapi"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/mailer"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/middleware0x"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool, sender mailer.Sender) http.Handler {
	csrf := http.NewCrossOriginProtection()
	if err := csrf.AddTrustedOrigin("http://localhost:5173"); err != nil {
		panic(err)
	}

	handlers := httpapi.NewHandler(pool, sender)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(csrf.Handler)

	// Routing GET requests
	r.Route("/api", func(r chi.Router) {
		r.HandleFunc("GET /health", httpapi.Health)

		r.Route("/profiles", func(r chi.Router) {
			r.HandleFunc("GET /{handle}", handlers.GetUserPublicProfile())
		})

		// Routing Authentication related
		r.Route("/auth", func(r chi.Router) {
			r.HandleFunc("POST /logout", handlers.Logout())
			r.HandleFunc("POST /register", handlers.Register())
			r.HandleFunc("POST /login", handlers.Login())
			r.HandleFunc("POST /verify-email", handlers.VerifyEmail())
			r.HandleFunc("POST /resend-verification", handlers.ResendEmailVerification())
			r.HandleFunc("POST /forgot-password", handlers.ForgotPassword())
			r.HandleFunc("POST /reset-password", handlers.PasswordReset())

			r.Route("/me", func(r chi.Router) {
				r.Use(middleware0x.RequireAuth(pool))

				r.HandleFunc("GET /profile", handlers.GetProfile())
				r.HandleFunc("PATCH /profile", handlers.UpdateProfile())
				r.HandleFunc("GET /sessions", handlers.ListSessions())
				r.HandleFunc("DELETE /sessions/{sessionID}", handlers.RevokeSession())
			})

		})
	})
	return r
}
