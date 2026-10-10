package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

const testToken = "secret"

func TestRequireMetricsToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		authHeader string
		wantStatus int
		wantCalled bool
	}{
		{"valid token", testToken, "Bearer " + testToken, http.StatusOK, true},
		{"no header", testToken, "", http.StatusUnauthorized, false},
		{"wrong token", testToken, "Bearer other", http.StatusUnauthorized, false},
		{"wrong scheme", testToken, "Basic secret", http.StatusUnauthorized, false},
		{"empty bearer value", testToken, "Bearer ", http.StatusUnauthorized, false},
		{"token not configured rejects empty bearer", "", "Bearer ", http.StatusUnauthorized, false},
		{"token not configured rejects any", "", "Bearer secret", http.StatusUnauthorized, false},
		{"token not configured rejects no header", "", "", http.StatusUnauthorized, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			h := middleware.RequireMetricsToken(tt.token)(next)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if called != tt.wantCalled {
				t.Errorf("next called = %v, want %v", called, tt.wantCalled)
			}
			if tt.wantStatus == http.StatusUnauthorized && rec.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Errorf("WWW-Authenticate = %q, want Bearer", rec.Header().Get("WWW-Authenticate"))
			}
		})
	}
}
