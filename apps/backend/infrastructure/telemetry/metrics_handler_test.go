package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/Haya372/ai-trial/backend/infrastructure/telemetry"
)

func TestMetricsHandler_returns200(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	handler := telemetry.NewMetricsHandler(reader)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMetricsHandler_outputsPrometheusContentType(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	handler := telemetry.NewMetricsHandler(reader)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("expected text/plain content type, got %q", ct)
	}
}

func TestMetricsHandler_outputsCounterMetric(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := mp.Meter("test")

	counter, err := meter.Int64Counter("http_requests_total")
	if err != nil {
		t.Fatal(err)
	}
	counter.Add(context.Background(), 5)

	handler := telemetry.NewMetricsHandler(reader)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Errorf("expected http_requests_total in output, got:\n%s", body)
	}
	if !strings.Contains(body, "5") {
		t.Errorf("expected value 5 in output, got:\n%s", body)
	}
}
