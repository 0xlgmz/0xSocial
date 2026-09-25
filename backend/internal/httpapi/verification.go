package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
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

		sessionToken, sessionTokenHash, err := auth.NewSessionToken()
		if err != nil {
			http.Error(w, "could not verify email", http.StatusInternalServerError)
			return
		}

		expiresAt, err := postgres.VerifyEmailAndCreateSession(
			r.Context(),
			h.pool,
			input.Token,
			sessionTokenHash,
			clientIP(r),
			r.UserAgent(),
		)
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

		http.SetCookie(w, &http.Cookie{
			Name:     "__Host-session",
			Value:    sessionToken,
			Path:     "/",
			Expires:  expiresAt,
			MaxAge:   int(time.Until(expiresAt).Seconds()),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		w.WriteHeader(http.StatusNoContent)
	}
}
func (h *Handler) ResendEmailVerification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if rateLimited(w, h.limiters.ResendVerification.IP, clientIP(r)) {
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

		if rateLimited(w, h.limiters.ResendVerification.Email, emailNormalized) {
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
				clientIP(r),
				r.UserAgent(),
			)

		if err != nil {
			http.Error(w, "could not process request", http.StatusInternalServerError)
			return
		}

		if shouldSend {
			if err := h.mailer.SendVerificationEmail(
				r.Context(),
				email,
				rawToken,
			); err != nil {
				slog.Error(
					"failed to resend verification email",
					"error", err,
				)
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
