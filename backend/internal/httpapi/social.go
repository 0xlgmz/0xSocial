package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/go-chi/chi/v5"
)

const (
	defaultPageSize = 20
	maximumPageSize = 50
	maximumPostSize = 2_000
)

type createPostRequest struct {
	Content  string  `json:"content"`
	MediaIDs []int64 `json:"mediaIds"`
}

type postResponse struct {
	ID          int64     `json:"id"`
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	AvatarURL   *string   `json:"avatarUrl"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Images      []string  `json:"images"`
}

type postsResponse struct {
	Posts      []postResponse `json:"posts"`
	NextCursor int64          `json:"nextCursor,omitempty"`
}

type followProfileResponse struct {
	Handle      string    `json:"handle"`
	DisplayName string    `json:"displayName"`
	Bio         string    `json:"bio"`
	AvatarURL   *string   `json:"avatarUrl"`
	FollowedAt  time.Time `json:"followedAt"`
}

type followProfilesResponse struct {
	Profiles   []followProfileResponse `json:"profiles"`
	NextCursor int64                   `json:"nextCursor,omitempty"`
}

type relationshipResponse struct {
	IsSelf     bool `json:"isSelf"`
	Following  bool `json:"following"`
	FollowedBy bool `json:"followedBy"`
}

func pagination(r *http.Request) (int64, int, error) {
	beforeID := int64(0)
	limit := defaultPageSize

	if raw := r.URL.Query().Get("before"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			return 0, 0, errors.New("invalid pagination cursor")
		}
		beforeID = parsed
	}

	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 || parsed > maximumPageSize {
			return 0, 0, errors.New("invalid page size")
		}
		limit = parsed
	}

	return beforeID, limit, nil
}

func requestedHandle(r *http.Request) string {
	return strings.ToLower(strings.TrimSpace(chi.URLParam(r, "handle")))
}

func (h *Handler) mapPost(post postgres.Post) postResponse {
	response := postResponse{
		ID:          post.ID,
		Handle:      post.Handle,
		DisplayName: post.DisplayName,
		AvatarURL:   post.AvatarURL,
		Content:     post.Content,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		Images:      make([]string, 0, len(post.MediaKeys)),
	}
	if h.media != nil {
		for _, objectKey := range post.MediaKeys {
			response.Images = append(response.Images, h.media.PublicURL(objectKey))
		}
	}

	return response
}

func (h *Handler) writePosts(w http.ResponseWriter, posts []postgres.Post) {
	response := postsResponse{
		Posts: make([]postResponse, 0, len(posts)),
	}
	for _, post := range posts {
		response.Posts = append(response.Posts, h.mapPost(post))
	}
	if len(posts) > 0 {
		response.NextCursor = posts[len(posts)-1].ID
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) CreatePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.CreatePost, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 16_384)
		var input createPostRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		content := strings.TrimSpace(input.Content)
		length := utf8.RuneCountInString(content)
		if length == 0 || length > maximumPostSize {
			http.Error(w, "post content must be between 1 and 2000 characters", http.StatusBadRequest)
			return
		}

		pendingMedia, err := h.verifiedPendingMedia(r.Context(), session.UserID, input.MediaIDs)
		if errors.Is(err, errInvalidUploadedMedia) {
			http.Error(w, "one or more images are invalid or have expired", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "could not verify uploaded images", http.StatusBadGateway)
			return
		}

		post, err := postgres.CreatePost(
			r.Context(), h.pool, session.UserID, content, pendingMedia,
		)
		if errors.Is(err, postgres.ErrMediaNotAttachable) {
			http.Error(w, "one or more images are no longer available", http.StatusConflict)
			return
		}
		if err != nil {
			http.Error(w, "could not create post", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(h.mapPost(post))
	}
}

func (h *Handler) DeletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.DeletePost, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
		if err != nil || postID <= 0 {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}

		objectKeys, err := postgres.DeletePost(r.Context(), h.pool, session.UserID, postID)
		if errors.Is(err, postgres.ErrPostNotFound) {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not delete post", http.StatusInternalServerError)
			return
		}

		if h.media != nil {
			for _, objectKey := range objectKeys {
				if err := h.media.DeleteObject(r.Context(), objectKey); err != nil {
					slog.Error(
						"failed to delete post image",
						"object_key", objectKey,
						"error", err,
					)
				}
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) ListProfilePosts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.SocialRead, clientIP(r)) {
			return
		}

		beforeID, limit, err := pagination(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		posts, err := postgres.ListPostsByHandle(
			r.Context(), h.pool, requestedHandle(r), beforeID, limit,
		)
		if errors.Is(err, postgres.ErrSocialProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not list posts", http.StatusInternalServerError)
			return
		}

		h.writePosts(w, posts)
	}
}

func (h *Handler) ListFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.SocialRead, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		beforeID, limit, err := pagination(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		posts, err := postgres.ListFollowingFeed(r.Context(), h.pool, session.UserID, beforeID, limit)
		if err != nil {
			http.Error(w, "could not load feed", http.StatusInternalServerError)
			return
		}

		h.writePosts(w, posts)
	}
}

func (h *Handler) ListExplore() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.SocialRead, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		beforeID, limit, err := pagination(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		posts, err := postgres.ListExplorePosts(r.Context(), h.pool, beforeID, limit)
		if err != nil {
			http.Error(w, "could not load explore posts", http.StatusInternalServerError)
			return
		}

		h.writePosts(w, posts)
	}
}

func (h *Handler) FollowUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.FollowMutation, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		err := postgres.FollowUser(r.Context(), h.pool, session.UserID, requestedHandle(r))
		if errors.Is(err, postgres.ErrSocialProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, postgres.ErrCannotFollowSelf) {
			http.Error(w, "you cannot follow yourself", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "could not follow profile", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) UnfollowUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.FollowMutation, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		err := postgres.UnfollowUser(r.Context(), h.pool, session.UserID, requestedHandle(r))
		if errors.Is(err, postgres.ErrSocialProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, postgres.ErrCannotFollowSelf) {
			http.Error(w, "you cannot unfollow yourself", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "could not unfollow profile", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *Handler) GetRelationship() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.SocialRead, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		relationship, err := postgres.GetRelationship(
			r.Context(), h.pool, session.UserID, requestedHandle(r),
		)
		if errors.Is(err, postgres.ErrSocialProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not load relationship", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(relationshipResponse{
			IsSelf:     relationship.IsSelf,
			Following:  relationship.Following,
			FollowedBy: relationship.FollowedBy,
		})
	}
}

func (h *Handler) ListFollowers() http.HandlerFunc {
	return h.listFollowProfiles(true)
}

func (h *Handler) ListFollowing() http.HandlerFunc {
	return h.listFollowProfiles(false)
}

func (h *Handler) listFollowProfiles(followers bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if rateLimited(w, h.limiters.SocialRead, clientIP(r)) {
			return
		}

		beforeID, limit, err := pagination(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var profiles []postgres.FollowProfile
		if followers {
			profiles, err = postgres.ListFollowers(
				r.Context(), h.pool, requestedHandle(r), beforeID, limit,
			)
		} else {
			profiles, err = postgres.ListFollowing(
				r.Context(), h.pool, requestedHandle(r), beforeID, limit,
			)
		}

		if errors.Is(err, postgres.ErrSocialProfileNotFound) {
			http.Error(w, "profile not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "could not list profiles", http.StatusInternalServerError)
			return
		}

		response := followProfilesResponse{
			Profiles: make([]followProfileResponse, 0, len(profiles)),
		}
		for _, profile := range profiles {
			response.Profiles = append(response.Profiles, followProfileResponse{
				Handle:      profile.Handle,
				DisplayName: profile.DisplayName,
				Bio:         profile.Bio,
				AvatarURL:   profile.AvatarURL,
				FollowedAt:  profile.FollowedAt,
			})
		}
		if len(profiles) > 0 {
			response.NextCursor = profiles[len(profiles)-1].Cursor
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}
