package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.Login.IP, clientIP(r)) {
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
		if rateLimited(w, h.limiters.Login.Email, emailNormalized) {
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
			h.recordAuthEventBestEffort(
				r,
				postgres.AuthEvent{
					EventType: postgres.AuthEventLoginFailed,
					Metadata: map[string]any{
						"reason": "invalid_credentials",
					},
				},
			)
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
			h.recordAuthEventBestEffort(
				r,
				postgres.AuthEvent{
					UserID:    &userID,
					EventType: postgres.AuthEventLoginFailed,
					Metadata: map[string]any{
						"reason": "invalid_credentials",
					},
				},
			)

			http.Error(w, "invalid email or password", http.StatusUnauthorized)
			return
		}

		if status != "active" {
			h.recordAuthEventBestEffort(
				r,
				postgres.AuthEvent{
					UserID:    &userID,
					EventType: postgres.AuthEventLoginFailed,
					Metadata: map[string]any{
						"reason": "account_not_active",
					},
				},
			)
			http.Error(w, "email verification required", http.StatusForbidden)
			return
		}

		h.limiters.Login.Email.Reset(emailNormalized)

		sessionToken, expiresAt, sessionID, err := generateSession(r, h.pool, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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

		h.recordAuthEventBestEffort(
			r,
			postgres.AuthEvent{
				UserID:    &userID,
				SessionID: &sessionID,
				EventType: postgres.AuthEventLoginSucceeded,
				Metadata: map[string]any{
					"method": "password",
				},
			},
		)

		w.WriteHeader(http.StatusOK)
	}
}

func generateSession(r *http.Request, pool *pgxpool.Pool, userID int64) (string, time.Time, int64, error) {
	sessionToken, sessionTokenHash, err := auth.NewSessionToken()
	if err != nil {
		return "", time.Time{}, 0, errors.New("could not generate session token")
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	var sessionID int64

	err = pool.QueryRow(
		r.Context(),
		`INSERT INTO sessions (
			user_id,
			token_hash,
			expires_at,
			user_agent,
			ip_address
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			NULLIF($5, '')::INET
		)
		RETURNING id`,
		userID,
		sessionTokenHash,
		expiresAt,
		r.UserAgent(),
		clientIP(r),
	).Scan(&sessionID)

	if err != nil {
		return "", time.Time{}, 0, errors.New("could not create session")
	}
	return sessionToken, expiresAt, sessionID, nil
}
