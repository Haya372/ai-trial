package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func TestHealthHandler_GET_returns200(t *testing.T) {
	h := handler.NewHealthHandler()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
