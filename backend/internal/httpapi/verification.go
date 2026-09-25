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
)

type emailVerification struct {
	Token string `json:"token"`
}
type resendVerificationRequest struct {
	Email string `json:"email"`
}

func (h *Handler) VerifyEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input emailVerification
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		err := postgres.VerifyEmailToken(r.Context(), h.pool, input.Token)
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
func (h *Handler) ResendEmailVerification(limiters ratelimit.IPAndEmailLimiters) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if rateLimited(w, limiters.IP, clientIP(r)) {
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

		if rateLimited(w, limiters.Email, emailNormalized) {
			return
		}

		rawToken, tokenHash, err := auth.NewOpaqueToken()
		if err != nil {
			http.Error(w, "could not process request", http.StatusInternalServerError)
			return
		}

		email, shouldSend, err :=
			postgres.ReplaceEmailVerificationToken(
				r.Context(),
				h.pool,
				emailNormalized,
				tokenHash,
			)

		if err != nil {
			http.Error(w, "could not process request", http.StatusInternalServerError)
			return
		}

		if shouldSend {
			// Development only. Replace with mail delivery later.
			log.Printf("verification email for %s: token=%s", email, rawToken)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handler) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (h *Handler) PasswordReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
