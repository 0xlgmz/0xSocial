package httpapi

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = pool // We will use this in the next stage.

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
		// HTTP Reponse for successful registration with hashed password (for demonstration purposes).
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"email":    emailNormalized,
			"password": hashedPassword,
		}
		json.NewEncoder(w).Encode(response)
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
	// Basic password validation logic.
	if len(password) < 15 {
		return false, "password must be at least 15 characters long"
	}
	// Check for maximum length to prevent abuse.
	if len(password) > 1_024 {
		return false, "password must not exceed 1024 characters"
	}
	// Check for at least one special character in the password.
	specialChar := regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`)
	if !specialChar.MatchString(password) {
		return false, "password must contain at least one special character (@, #, $, %)"
	}
	return true, ""
}
