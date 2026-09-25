package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/ratelimit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(limiters ratelimit.IPAndEmailLimiters) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, limiters.IP, clientIP(r)) {
			return
		}
		var (
			userID       int64
			passwordHash string
			status       string
		)
		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input loginRequest
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

		err := h.pool.QueryRow(
			r.Context(),
			`SELECT u.id, pc.password_hash, u.status
			FROM users u
			JOIN password_credentials pc ON pc.user_id = u.id
			WHERE u.email_normalized = $1`,
			emailNormalized,
		).Scan(&userID, &passwordHash, &status)

		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}

		if err != nil {
			http.Error(w, "could not log in", http.StatusInternalServerError)
			return
		}

		matches, err := auth.VerifyPassword(input.Password, passwordHash)
		if err != nil {
			http.Error(w, "could not log in", http.StatusInternalServerError)
			return
		}

		if !matches {
			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}

		if status != "active" {
			http.Error(w, "email verification required", http.StatusForbidden)
			return
		}

		limiters.Email.Reset(emailNormalized)

		generateErr, sessionToken, expiresAt := generateSession(r, h.pool, userID)
		if generateErr != nil {
			http.Error(w, generateErr.Error(), http.StatusInternalServerError)
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

		w.WriteHeader(http.StatusOK)
	}
}

func generateSession(r *http.Request, pool *pgxpool.Pool, userID int64) (error, string, time.Time) {
	sessionToken, sessionTokenHash, err := auth.NewSessionToken()
	if err != nil {
		return errors.New("could not generate session token"), "", time.Time{}
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err = pool.Exec(
		r.Context(),
		`INSERT INTO sessions (user_id, token_hash, expires_at, user_agent)
		VALUES ($1, $2, $3, $4)`,
		userID,
		sessionTokenHash,
		expiresAt,
		r.UserAgent(),
	)

	if err != nil {
		return errors.New("could not create session"), "", time.Time{}
	}
	return nil, sessionToken, expiresAt
}
