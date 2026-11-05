package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	sessionmock "github.com/Haya372/ai-trial/backend/domain/session/generated"
	usermock "github.com/Haya372/ai-trial/backend/domain/user/generated"
	"github.com/Haya372/ai-trial/backend/interface/middleware"
	"github.com/Haya372/ai-trial/backend/interface/route"
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
	if v, _ := attrs["method"].(string); v != http.MethodGet {
		t.Errorf("expected method %q, got %q", http.MethodGet, v)
	}
	if v, _ := attrs["path"].(string); v != "/events" {
		t.Errorf("expected path %q, got %q", "/events", v)
	}
}

func TestAccessLog_noWriteHeader_treatsAs200(t *testing.T) {
	h := &captureHandler{}
	logger := slog.New(h)
	noWriteHeader := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {})
	mw := middleware.AccessLog(logger)(noWriteHeader)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req)

	if len(h.records) != 1 {
		t.Fatalf("expected 1 log record, got %d", len(h.records))
	}
	attrs := loggedAttrs(h.records[0])
	if v, _ := attrs["status"].(int64); v != http.StatusOK {
		t.Errorf("expected status 200, got %v", v)
	}
	if h.records[0].Level != slog.LevelInfo {
		t.Errorf("expected Info level, got %s", h.records[0].Level)
	}
}

const secretShareToken = "s3cr3t-share-token-value"

// newShareRouter wires the share routes the same way di.container does, so
// the test exercises the real chi route context the middlewares log from.
func newShareRouter(t *testing.T, logger *slog.Logger) http.Handler {
	t.Helper()
	ctrl := gomock.NewController(t)
	sessRepo := sessionmock.NewMockRepository(ctrl)
	sessRepo.EXPECT().FindActiveByTokenHash(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	userRepo := usermock.NewMockRepository(ctrl)

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	r := chi.NewRouter()
	r.Use(middleware.AccessLog(logger))
	r.With(middleware.OptionalAuth(sessRepo, userRepo, logger)).Get("/shares/{token}", ok)
	r.With(middleware.RequireAuth(sessRepo, userRepo, logger)).Post("/shares/{token}/subscriptions", ok)
	return r
}

func TestAccessLog_shareRoutes_doNotLogShareToken(t *testing.T) {
	cases := []struct {
		name        string
		method      string
		path        string
		sessionID   string
		wantPattern string
	}{
		{"GET share without cookie", http.MethodGet, "/shares/" + secretShareToken, "", "/shares/{token}"},
		{
			"GET share with expired session", http.MethodGet,
			"/shares/" + secretShareToken, uuid.NewString(), "/shares/{token}",
		},
		{
			"POST subscription without cookie", http.MethodPost,
			"/shares/" + secretShareToken + "/subscriptions", "", "/shares/{token}/subscriptions",
		},
		{
			"POST subscription with expired session", http.MethodPost,
			"/shares/" + secretShareToken + "/subscriptions", uuid.NewString(), "/shares/{token}/subscriptions",
		},
		{"DELETE share (method not allowed)", http.MethodDelete, "/shares/" + secretShareToken, "", route.Unmatched},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			router := newShareRouter(t, logger)

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			if tc.sessionID != "" {
				req.AddCookie(&http.Cookie{Name: testSessionCookieName, Value: tc.sessionID}) //nolint:gosec
			}
			router.ServeHTTP(httptest.NewRecorder(), req)

			if strings.Contains(buf.String(), secretShareToken) {
				t.Fatalf("share token leaked into logs:\n%s", buf.String())
			}
			assertAccessLogPath(t, buf.String(), tc.wantPattern)
		})
	}
}

func assertAccessLogPath(t *testing.T, logs, want string) {
	t.Helper()
	for line := range strings.SplitSeq(strings.TrimSpace(logs), "\n") {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		if rec["msg"] != "access" {
			continue
		}
		if rec["path"] != want {
			t.Errorf("expected access log path %q, got %v", want, rec["path"])
		}
		return
	}
	t.Fatalf("access log record not found in:\n%s", logs)
}
