package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
)

type emailVerification struct {
	Token string `json:"token"`
}
type resendVerificationRequest struct {
	Email string `json:"email"`
}

func VerifyEmail(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input emailVerification
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		err := postgres.VerifyEmailToken(r.Context(), pool, input.Token)
		if errors.Is(err, postgres.ErrInvalidVerificationToken) {
			http.Error(
				w,
				"invalid or expired verification token",
				http.StatusBadRequest,
			)
			return
		}

		if err != nil {
			http.Error(w, "could not verify email", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ResendEmailVerification(pool *pgxpool.Pool, resendIPLimiter, resendEmailLimiter *ratelimit.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if rateLimited(w, resendIPLimiter, clientIP(r)) {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input resendVerificationRequest

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		emailNormalized := strings.ToLower(strings.TrimSpace(input.Email))

		if rateLimited(w, resendEmailLimiter, emailNormalized) {
			return
		}

		rawToken, tokenHash, err := auth.NewOpaqueToken()
		if err != nil {
			http.Error(w, "could not process request", http.StatusInternalServerError)
			return
		}

		if err != nil {
			http.Error(w, "could not process request", http.StatusInternalServerError)
			return
		}

		log.Println("raw: [%s] \ntokenHash: [%s]", rawToken, tokenHash)
		// Always return the same response.
		w.WriteHeader(http.StatusAccepted)

	}
}
