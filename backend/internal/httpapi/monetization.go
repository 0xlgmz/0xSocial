package httpapi

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) GetMonetizationConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(h.monetization)
	}
}
