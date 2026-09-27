package httpapi

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/0xlgmz/proj-reactlang-fullstack/internal/monetization"
)

func TestGetMonetizationConfigDisabled(t *testing.T) {
	handler := &Handler{monetization: monetization.DisabledConfig()}
	request := httptest.NewRequest("GET", "/api/monetization/config", nil)
	response := httptest.NewRecorder()

	handler.GetMonetizationConfig().ServeHTTP(response, request)

	if response.Code != 200 {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}

	var config monetization.Config
	if err := json.NewDecoder(response.Body).Decode(&config); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if config.Enabled || config.Placements.Feed.Enabled {
		t.Fatalf("config = %+v, want disabled", config)
	}
}
