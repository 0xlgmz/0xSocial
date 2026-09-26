package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/go-chi/chi/v5"
)

type profileResponse struct {
	ID               int64     `json:"id"`
	Email            string    `json:"email"`
	Status           string    `json:"status"`
	Handle           string    `json:"handle"`
	DisplayName      string    `json:"displayName"`
	Bio              string    `json:"bio"`
	AvatarURL        *string   `json:"avatarUrl"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	SessionExpiresAt time.Time `json:"sessionExpiresAt"`
}
type updateProfileRequest struct {
	DisplayName *string `json:"displayName"`
	Bio         *string `json:"bio"`
}
type publicProfileResponse struct {
	Handle      string  `json:"handle"`
	DisplayName string  `json:"displayName"`
	Bio         string  `json:"bio"`
	AvatarURL   *string `json:"avatarUrl"`
}

func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.GetProfile, clientIP(r)) {
			return
		}

		profile, err := postgres.GetUserProfile(
			r.Context(),
			h.pool,
			session.UserID,
		)
		if err != nil {
			http.Error(w, "could not load profile", http.StatusInternalServerError)
			return
		}

		response := profileResponse{
			ID:               session.UserID,
			Email:            session.Email,
			Status:           session.Status,
			Handle:           profile.Handle,
			DisplayName:      profile.DisplayName,
			Bio:              profile.Bio,
			AvatarURL:        profile.AvatarURL,
			CreatedAt:        profile.CreatedAt,
			UpdatedAt:        profile.UpdatedAt,
			SessionExpiresAt: session.ExpiresAt,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
		h.limiters.GetProfile.Reset(clientIP(r))
	}
}
func (h *Handler) UpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.UpdateProfile, clientIP(r)) {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4_096)

		var input updateProfileRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if input.DisplayName == nil && input.Bio == nil {
			http.Error(w, "at least one profile field is required", http.StatusBadRequest)
			return
		}

		if input.DisplayName != nil {
			normalized := strings.TrimSpace(*input.DisplayName)

			if utf8.RuneCountInString(normalized) > 80 {
				http.Error(w, "display name must not exceed 80 characters", http.StatusBadRequest)
				return
			}

			input.DisplayName = &normalized
		}

		if input.Bio != nil {
			normalized := strings.TrimSpace(*input.Bio)

			if utf8.RuneCountInString(normalized) > 500 {
				http.Error(w, "bio must not exceed 500 characters", http.StatusBadRequest)
				return
			}

			input.Bio = &normalized
		}

		profile, err := postgres.UpdateUserProfile(
			r.Context(),
			h.pool,
			session.UserID,
			input.DisplayName,
			input.Bio,
		)
		if err != nil {
			http.Error(w, "could not update profile", http.StatusInternalServerError)
			return
		}

		response := profileResponse{
			ID:               session.UserID,
			Email:            session.Email,
			Status:           session.Status,
			Handle:           profile.Handle,
			DisplayName:      profile.DisplayName,
			Bio:              profile.Bio,
			AvatarURL:        profile.AvatarURL,
			CreatedAt:        profile.CreatedAt,
			UpdatedAt:        profile.UpdatedAt,
			SessionExpiresAt: session.ExpiresAt,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
		h.limiters.GetProfile.Reset(clientIP(r))
	}
}

func (h *Handler) GetUserPublicProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.GetProfile, clientIP(r)) {
			return
		}
		profile, err := postgres.GetUserPublicProfile(
			r.Context(),
			h.pool,
			chi.URLParam(r, "handle"),
		)
		if errors.Is(err, postgres.ErrPublicProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not load profile", http.StatusInternalServerError)
			return
		}

		response := publicProfileResponse{
			Handle:      profile.Handle,
			DisplayName: profile.DisplayName,
			Bio:         profile.Bio,
			AvatarURL:   profile.AvatarURL,
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(response); err != nil {
			return
		}
		h.limiters.GetProfile.Reset(clientIP(r))
	}
}
