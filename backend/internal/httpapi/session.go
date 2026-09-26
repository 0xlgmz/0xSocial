package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/middleware0x"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/go-chi/chi/v5"
)

type listedSessionResponse struct {
	ID         int64     `json:"id"`
	Current    bool      `json:"current"`
	UserAgent  string    `json:"userAgent"`
	IPAddress  string    `json:"ipAddress"`
	AuthMethod string    `json:"authMethod"`
	RiskLevel  string    `json:"riskLevel"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type listSessionsResponse struct {
	Sessions []listedSessionResponse `json:"sessions"`
}

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

func (h *Handler) ListSessions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentSession, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.GetProfile, clientIP(r)) {
			return
		}

		sessions, err := postgres.ListUserSessions(
			r.Context(),
			h.pool,
			currentSession.UserID,
		)
		if err != nil {
			http.Error(w, "could not list sessions", http.StatusInternalServerError)
			return
		}

		response := listSessionsResponse{
			Sessions: make([]listedSessionResponse, 0, len(sessions)),
		}

		for _, session := range sessions {
			response.Sessions = append(
				response.Sessions,
				listedSessionResponse{
					ID:         session.ID,
					Current:    session.ID == currentSession.SessionID,
					UserAgent:  session.UserAgent,
					IPAddress:  session.IPAddress,
					AuthMethod: session.AuthMethod,
					RiskLevel:  session.RiskLevel,
					CreatedAt:  session.CreatedAt,
					LastSeenAt: session.LastSeenAt,
					ExpiresAt:  session.ExpiresAt,
				},
			)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
		h.limiters.GetProfile.Reset(clientIP(r))
	}
}
func (h *Handler) RevokeSession() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentSession, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.DeleteSession, clientIP(r)) {
			return
		}

		targetSessionID, err := strconv.ParseInt(
			chi.URLParam(r, "sessionID"),
			10,
			64,
		)
		if err != nil || targetSessionID <= 0 {
			http.Error(w, "invalid session id", http.StatusBadRequest)
			return
		}

		revoked, err := postgres.RevokeUserSession(
			r.Context(),
			h.pool,
			currentSession.UserID,
			targetSessionID,
			currentSession.SessionID,
			clientIP(r),
			r.UserAgent(),
		)
		if err != nil {
			http.Error(w, "could not revoke session", http.StatusInternalServerError)
			return
		}

		if !revoked {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		// If the user revoked their current session, remove its cookie.
		if targetSessionID == currentSession.SessionID {
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
		}

		h.limiters.DeleteSession.Reset(clientIP(r))

		w.WriteHeader(http.StatusNoContent)
	}
}
