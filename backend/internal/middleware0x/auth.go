package middleware0x

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthContextKey struct{}

type AuthenticatedSession struct {
	UserID     int64
	SessionID  int64
	Email      string
	Status     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type SessionResponse struct {
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func RequireAuth(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("__Host-session")
			if err != nil {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}

			var session AuthenticatedSession

			err = pool.QueryRow(
				r.Context(),
				`SELECT u.id, s.id, u.email, u.status, s.expires_at, s.created_at, s.last_seen_at
				 FROM sessions s
				 JOIN users u ON u.id = s.user_id
				 WHERE s.token_hash = $1
				   AND s.revoked_at IS NULL
				   AND s.expires_at > NOW()
				   AND s.created_at >= u.sessions_valid_after
				   AND u.status IN ('pending_verification', 'active')
				   AND s.created_at + INTERVAL '30 days' > NOW()`,
				auth.HashSessionToken(cookie.Value),
			).Scan(
				&session.UserID,
				&session.SessionID,
				&session.Email,
				&session.Status,
				&session.ExpiresAt,
			)

			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}

			if err != nil {
				http.Error(w, "could not authenticate request", http.StatusInternalServerError)
				return
			}

			if time.Since(session.LastSeenAt) >= 15*time.Minute {
				var refreshedExpiry time.Time

				err := pool.QueryRow(
					r.Context(),
					`UPDATE sessions
					SET last_seen_at = NOW(),
						expires_at = LEAST(
							NOW() + INTERVAL '7 days',
							created_at + INTERVAL '30 days'
						)
					WHERE id = $1
					RETURNING expires_at`,
					session.SessionID,
				).Scan(&refreshedExpiry)

				if err != nil {
					http.Error(w, "could not refresh session", http.StatusInternalServerError)
					return
				}

				session.ExpiresAt = refreshedExpiry

				http.SetCookie(w, &http.Cookie{
					Name:     "__Host-session",
					Value:    cookie.Value,
					Path:     "/",
					Expires:  refreshedExpiry,
					MaxAge:   int(time.Until(refreshedExpiry).Seconds()),
					Secure:   true,
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})
			}

			ctx := context.WithValue(
				r.Context(),
				AuthContextKey{},
				session,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
