package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

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

func TestHealthRouter_nonGET_returns405(t *testing.T) {
	h := handler.NewHealthHandler()
	r := chi.NewRouter()
	r.Get("/health", h.ServeHTTP)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}
