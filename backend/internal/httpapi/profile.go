package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type profileResponse struct {
	ID               int64     `json:"id"`
	Email            string    `json:"email"`
	Status           string    `json:"status"`
	SessionExpiresAt time.Time `json:"sessionExpiresAt"`
}

func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}

		response := profileResponse{
			ID:               session.UserID,
			Email:            session.Email,
			Status:           session.Status,
			SessionExpiresAt: session.ExpiresAt,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
	}
}
func (h *Handler) UpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "profile updates are not implemented", http.StatusNotImplemented)
	}
}
