package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/postgres"
	"github.com/go-chi/chi/v5"
)

type reportPostRequest struct {
	Reason string `json:"reason"`
}

var reportReasons = map[string]struct{}{
	"spam":              {},
	"harassment":        {},
	"hate_speech":       {},
	"violence":          {},
	"sexual_content":    {},
	"self_harm":         {},
	"false_information": {},
	"other":             {},
}

func validReportReason(reason string) bool {
	_, ok := reportReasons[reason]
	return ok
}

func (h *Handler) ReportPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := sessionFromContext(r)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if rateLimited(w, h.limiters.ReportPost, strconv.FormatInt(session.UserID, 10)) {
			return
		}

		postID, err := strconv.ParseInt(chi.URLParam(r, "postID"), 10, 64)
		if err != nil || postID <= 0 {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 4_096)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()

		var input reportPostRequest
		if err := decoder.Decode(&input); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if !validReportReason(input.Reason) {
			http.Error(w, "invalid report reason", http.StatusBadRequest)
			return
		}

		err = postgres.CreatePostReport(r.Context(), h.pool, session.UserID, postID, input.Reason)
		switch {
		case errors.Is(err, postgres.ErrPostNotFound):
			http.Error(w, "post not found", http.StatusNotFound)
		case errors.Is(err, postgres.ErrCannotReportOwnPost):
			http.Error(w, "you cannot report your own post", http.StatusBadRequest)
		case errors.Is(err, postgres.ErrDuplicatePostReport):
			http.Error(w, "you have already reported this post", http.StatusConflict)
		case err != nil:
			http.Error(w, "could not report post", http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
