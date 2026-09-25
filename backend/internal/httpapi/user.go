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

type forgotPasswordRequest struct {
	Email string `json:"email"`
}
type passwordResetRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (h *Handler) ForgotPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.ForgotPassword.IP, clientIP(r)) {
			return
		}
		var request forgotPasswordRequest

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		email := strings.ToLower(strings.TrimSpace(request.Email))

		if rateLimited(w, h.limiters.ForgotPassword.Email, email) {
			return
		}

		rawToken, tokenHash, err := auth.NewOpaqueToken()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		shouldSend, err := postgres.ReplacePasswordResetToken(
			r.Context(),
			h.pool,
			email,
			tokenHash,
		)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if shouldSend {
			if err := h.mailer.SendPasswordResetEmail(
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
func (h *Handler) PasswordReset() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.PasswordReset, clientIP(r)) {
			return
		}
		var request passwordResetRequest

		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&request); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if validPassword, passwordError := auth.ValidatePassword(request.Password); !validPassword {
			http.Error(w, passwordError, http.StatusBadRequest)
			return
		}

		passwordHash, err := auth.HashPassword(request.Password)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		email, err := postgres.ResetPassword(r.Context(), h.pool, request.Token, passwordHash)
		if errors.Is(err, postgres.ErrInvalidPasswordResetToken) {
			http.Error(
				w,
				"invalid or expired password reset token",
				http.StatusBadRequest,
			)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "__Host-session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(1, 0),
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})

		if err := h.mailer.SendPasswordChangedEmail(
			r.Context(),
			email,
		); err != nil {
			slog.Error(
				"failed to send password changed notification",
				"error", err,
			)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
