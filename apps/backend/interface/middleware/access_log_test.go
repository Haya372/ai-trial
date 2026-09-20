package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func handlerWithStatus(status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
}

func loggedAttrs(r slog.Record) map[string]any {
	m := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		m[a.Key] = a.Value.Any()
		return true
	})
	return m
}

func TestAccessLog_200_logsAtInfoLevel(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)
	mw := middleware.AccessLog(logger)(handlerWithStatus(http.StatusOK))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if len(h.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(h.records))
	}
	if h.records[0].Level != slog.LevelInfo {
		t.Errorf("expected Info level, got %s", h.records[0].Level)
	}
}

func TestAccessLog_400_logsAtWarnLevel(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)
	mw := middleware.AccessLog(logger)(handlerWithStatus(http.StatusBadRequest))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/login", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if len(h.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(h.records))
	}
	if h.records[0].Level != slog.LevelWarn {
		t.Errorf("expected Warn level, got %s", h.records[0].Level)
	}
}

func TestAccessLog_500_logsAtErrorLevel(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)
	mw := middleware.AccessLog(logger)(handlerWithStatus(http.StatusInternalServerError))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if len(h.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(h.records))
	}
	if h.records[0].Level != slog.LevelError {
		t.Errorf("expected Error level, got %s", h.records[0].Level)
	}
}

func TestAccessLog_logsRequiredFields(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)
	mw := middleware.AccessLog(logger)(handlerWithStatus(http.StatusOK))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if len(h.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(h.records))
	}
	attrs := loggedAttrs(h.records[0])
	for _, key := range []string{"method", "path", "status", "latency", "remote_addr"} {
		if _, ok := attrs[key]; !ok {
			t.Errorf("expected log field %q to be present", key)
		}
	}
}
