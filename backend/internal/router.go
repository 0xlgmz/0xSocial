package internal

import (
	"net/http"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/httpapi"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/middleware0x"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
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
	registrationLimiter := ratelimit.New(10, time.Hour)
	loginIPLimiter := ratelimit.New(20, time.Minute)
	loginEmailLimiter := ratelimit.New(5, 15*time.Minute)
	resendIPLimiter := ratelimit.New(5, time.Hour)
	resendEmailLimiter := ratelimit.New(3, time.Hour)

	// Routing GET requests
	r.Route("/api", func(r chi.Router) {
		r.HandleFunc("GET /health", httpapi.Health)

		// Routing Authentication related
		r.Route("/auth", func(r chi.Router) {
			r.HandleFunc("POST /logout", httpapi.Logout(pool))
			r.HandleFunc("POST /register", httpapi.Register(pool, registrationLimiter))
			r.HandleFunc("POST /login", httpapi.Login(pool, loginIPLimiter, loginEmailLimiter))
			r.HandleFunc("POST /verify-email", httpapi.VerifyEmail(pool))
			r.HandleFunc("POST /resend-verification", httpapi.ResendEmailVerification(pool, resendIPLimiter, resendEmailLimiter))

			r.Route("/me", func(r chi.Router) {
				r.Use(middleware0x.RequireAuth(pool))

				r.HandleFunc("GET /profile", httpapi.GetProfile(pool))
				r.HandleFunc("PATCH /profile", httpapi.UpdateProfile(pool))
				r.HandleFunc("GET /sessions", httpapi.ListSessions(pool))
				r.HandleFunc("DELETE /sessions/{sessionID}", httpapi.RevokeSession(pool))
			})

		})
	})
	return r
}
