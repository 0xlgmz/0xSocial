package httpapi

import (
	"net/http"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
)

func (h *Handler) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("__Host-session")

		if err == nil {
			_, err = h.pool.Exec(
				r.Context(),
				`UPDATE sessions
				 SET revoked_at = NOW(),
				     revoked_reason = 'user_logout'
				 WHERE token_hash = $1
				   AND revoked_at IS NULL`,
				auth.HashSessionToken(cookie.Value),
			)

			if err != nil {
				http.Error(w, "could not log out", http.StatusInternalServerError)
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "__Host-session",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(1, 0),
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		w.WriteHeader(http.StatusNoContent)
	}
}
