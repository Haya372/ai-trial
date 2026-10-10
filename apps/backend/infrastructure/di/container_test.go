// package di (not di_test): newRouter is unexported wiring with no public API
// surface worth exporting solely for tests.
//
//nolint:testpackage
package di

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const routerTestToken = "router-test-token"

// newRouter is built with nil handlers: the /metrics route is the only one
// exercised, so no database or use case is needed.
func newMetricsTestRouter(t *testing.T, token string) http.Handler {
	t.Helper()
	t.Setenv("METRICS_BEARER_TOKEN", token)
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return newRouter(
		nil, nil, nil, nil, nil, nil,
		nil, nil,
		slog.New(slog.DiscardHandler),
		tracenoop.NewTracerProvider(),
		mp,
		reader,
	)
}

func TestNewRouter_metricsRequiresBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		authHeader string
		wantStatus int
	}{
		{"valid token", routerTestToken, "Bearer " + routerTestToken, http.StatusOK},
		{"no header", routerTestToken, "", http.StatusUnauthorized},
		{"wrong token", routerTestToken, "Bearer other", http.StatusUnauthorized},
		{"token not configured", "", "Bearer " + routerTestToken, http.StatusUnauthorized},
		{"token not configured and empty bearer", "", "Bearer ", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newMetricsTestRouter(t, tt.token)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
