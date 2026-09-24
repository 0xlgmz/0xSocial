package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
	"github.com/jackc/pgx/v5/pgxpool"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(pool *pgxpool.Pool, registrationLimiter *ratelimit.Limiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, registrationLimiter, clientIP(r)) {
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
		validEmail := validateEmail(emailNormalized)
		if !validEmail {
			http.Error(w, "invalid email format", http.StatusBadRequest)
			return
		}
		validPassword, passwordError := validatePassword(input.Password)
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
			pool,
			emailNormalized,
			hashedPassword,
			verificationTokenHash,
		)

		if errors.Is(err, postgres.ErrEmailAlreadyExists) {
			http.Error(w, "an account with that email already exists", http.StatusConflict)
			return
		}

		if err != nil {
			http.Error(w, "could not create account", http.StatusInternalServerError)
			return
		}

		log.Printf("New User: [%s] \n Token: [%s]", emailNormalized, verificationToken)
		w.WriteHeader(http.StatusCreated)
	}
}

func validateEmail(email string) bool {
	// Basic email validation logic.
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	// Check for the presence of '@' symbol in the email.
	if !strings.Contains(email, "@") {
		return false
	}
	return true
}
func validatePassword(password string) (bool, string) {
	if len(password) > 1_024 {
		return false, "password must not exceed 1024 bytes"
	}

	length := utf8.RuneCountInString(password)

	if length < 8 {
		return false, "password must be at least 8 characters long"
	}

	// Long passwords/passphrases need no composition rules.
	if length >= 15 {
		return true, ""
	}

	var hasUppercase bool
	var hasSymbol bool

	for _, character := range password {
		if unicode.IsUpper(character) {
			hasUppercase = true
		}

		if unicode.IsPunct(character) || unicode.IsSymbol(character) {
			hasSymbol = true
		}
	}

	if !hasUppercase || !hasSymbol {
		return false, "passwords shorter than 15 characters require an uppercase letter and a symbol"
	}

	return true, ""
}
