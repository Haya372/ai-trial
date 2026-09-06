package handler_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func TestHealthHandler_GET_returns200(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := handler.NewHealthHandler(logger)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
