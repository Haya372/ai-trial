package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func TestHealth_GET_returns200(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
