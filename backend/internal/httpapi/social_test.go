package httpapi

import (
	"net/http/httptest"
	"testing"
)

func TestPagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantBefore int64
		wantLimit  int
		wantError  bool
	}{
		{name: "defaults", wantLimit: defaultPageSize},
		{name: "valid cursor and limit", query: "?before=42&limit=10", wantBefore: 42, wantLimit: 10},
		{name: "maximum limit", query: "?limit=50", wantLimit: maximumPageSize},
		{name: "zero cursor", query: "?before=0", wantError: true},
		{name: "negative cursor", query: "?before=-1", wantError: true},
		{name: "non-numeric cursor", query: "?before=post", wantError: true},
		{name: "zero limit", query: "?limit=0", wantError: true},
		{name: "limit too large", query: "?limit=51", wantError: true},
		{name: "non-numeric limit", query: "?limit=many", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/api/explore"+test.query, nil)
			before, limit, err := pagination(request)
			if test.wantError {
				if err == nil {
					t.Fatal("pagination() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("pagination() error = %v", err)
			}
			if before != test.wantBefore || limit != test.wantLimit {
				t.Fatalf("pagination() = (%d, %d), want (%d, %d)", before, limit, test.wantBefore, test.wantLimit)
			}
		})
	}
}
