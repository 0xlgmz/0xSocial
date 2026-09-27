package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/auth"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/media"
	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
)

const (
	maximumImageSize  = 10 * 1024 * 1024
	maximumPostImages = 4
	uploadURLLifetime = 5 * time.Minute
)

var errInvalidUploadedMedia = errors.New("uploaded media is invalid")

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type createMediaUploadRequest struct {
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type createMediaUploadResponse struct {
	MediaID   int64             `json:"mediaId"`
	UploadURL string            `json:"uploadUrl"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

func (h *Handler) CreateMediaUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if h.media == nil {
			http.Error(w, "media uploads are unavailable", http.StatusServiceUnavailable)
			return
		}
		if rateLimited(w, h.limiters.MediaUpload, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4_096)
		var input createMediaUploadRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		extension, supported := imageExtensions[input.ContentType]
		if !supported {
			http.Error(w, "unsupported image type", http.StatusBadRequest)
			return
		}
		if input.SizeBytes <= 0 || input.SizeBytes > maximumImageSize {
			http.Error(w, "image must not exceed 10 MB", http.StatusBadRequest)
			return
		}

		randomKey, _, err := auth.NewOpaqueToken()
		if err != nil {
			http.Error(w, "could not create upload", http.StatusInternalServerError)
			return
		}
		objectKey := fmt.Sprintf("posts/%d/%s%s", session.UserID, randomKey, extension)

		pending, err := postgres.CreatePendingMedia(
			r.Context(),
			h.pool,
			session.UserID,
			objectKey,
			input.ContentType,
			input.SizeBytes,
		)
		if err != nil {
			http.Error(w, "could not create upload", http.StatusInternalServerError)
			return
		}

		grant, err := h.media.PresignUpload(
			r.Context(), pending.ObjectKey, pending.MimeType, uploadURLLifetime,
		)
		if err != nil {
			_ = postgres.DeletePendingMedia(r.Context(), h.pool, session.UserID, pending.ID)
			http.Error(w, "could not create upload", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(createMediaUploadResponse{
			MediaID:   pending.ID,
			UploadURL: grant.URL,
			Method:    http.MethodPut,
			Headers: map[string]string{
				"Content-Type":  pending.MimeType,
				"If-None-Match": "*",
			},
			ExpiresAt: grant.ExpiresAt,
		})
	}
}

func (h *Handler) verifiedPendingMedia(
	ctx context.Context,
	userID int64,
	mediaIDs []int64,
) ([]postgres.PendingMedia, error) {
	if len(mediaIDs) == 0 {
		return []postgres.PendingMedia{}, nil
	}
	if h.media == nil || len(mediaIDs) > maximumPostImages {
		return nil, errInvalidUploadedMedia
	}

	seen := make(map[int64]struct{}, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		if mediaID <= 0 {
			return nil, errInvalidUploadedMedia
		}
		if _, duplicate := seen[mediaID]; duplicate {
			return nil, errInvalidUploadedMedia
		}
		seen[mediaID] = struct{}{}
	}

	pending, err := postgres.GetPendingMedia(ctx, h.pool, userID, mediaIDs)
	if errors.Is(err, postgres.ErrMediaNotAttachable) {
		return nil, errInvalidUploadedMedia
	}
	if err != nil {
		return nil, err
	}

	for _, item := range pending {
		object, err := h.media.HeadObject(ctx, item.ObjectKey)
		if errors.Is(err, media.ErrObjectNotFound) {
			return nil, errInvalidUploadedMedia
		}
		if err != nil {
			return nil, err
		}
		if object.ContentType != item.MimeType ||
			object.SizeBytes != item.SizeBytes ||
			object.SizeBytes <= 0 ||
			object.SizeBytes > maximumImageSize {
			return nil, errInvalidUploadedMedia
		}
	}

	return pending, nil
}
