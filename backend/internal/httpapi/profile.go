package httpapi

import (
	"net/http"
)

func (h *Handler) GetProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
func (h *Handler) UpdateProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
