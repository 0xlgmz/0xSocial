package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/middleware0x"
)

func CurrentSession(w http.ResponseWriter, r *http.Request) {
	session, ok := sessionFromContext(r)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	response := middleware0x.SessionResponse{
		Email:     session.Email,
		Status:    session.Status,
		ExpiresAt: session.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}
func sessionFromContext(r *http.Request) (middleware0x.AuthenticatedSession, bool) {
	session, ok := r.Context().Value(middleware0x.AuthContextKey{}).(middleware0x.AuthenticatedSession)
	return session, ok
}
