package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.Registration, clientIP(r)) {
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input registerRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		emailNormalized := strings.ToLower(strings.TrimSpace(input.Email))

		// Your validation goes here.
		validEmail := auth.ValidateEmail(emailNormalized)
		if !validEmail {
			http.Error(w, "invalid email format", http.StatusBadRequest)
			return
		}
		validPassword, passwordError := auth.ValidatePassword(input.Password)
		if !validPassword {
			http.Error(w, passwordError, http.StatusBadRequest)
			return
		}

		// Hash the password and store the user in the database.
		hashedPassword, err := auth.HashPassword(input.Password)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}

		verificationToken, verificationTokenHash, err := auth.NewOpaqueToken()
		if err != nil {
			http.Error(w, "could not create account", http.StatusInternalServerError)
			return
		}

		// Append to the database atomically.
		err = postgres.InsertRegisteredUser(
			r.Context(),
			h.pool,
			emailNormalized,
			hashedPassword,
			verificationTokenHash,
			clientIP(r),
			r.UserAgent(),
		)

		if errors.Is(err, postgres.ErrEmailAlreadyExists) {
			http.Error(w, "an account with that email already exists", http.StatusConflict)
			return
		}

		if err != nil {
			http.Error(w, "could not create account", http.StatusInternalServerError)
			return
		}

		if err := h.mailer.SendVerificationEmail(
			r.Context(),
			emailNormalized,
			verificationToken,
		); err != nil {
			slog.Error(
				"failed to send verification email",
				"error", err,
			)
		}
		w.WriteHeader(http.StatusCreated)
	}
}
