package telemetry_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

func TestMetricsHandler_histogramInfBucket_isValidPrometheusFormat(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := mp.Meter("test")

	hist, err := meter.Float64Histogram("http_request_duration_seconds")
	if err != nil {
		t.Fatal(err)
	}
	hist.Record(context.Background(), 0.1,
		metric.WithAttributes(attribute.String("method", "GET"), attribute.String("path", "/health")))

	handler := telemetry.NewMetricsHandler(reader)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()
	// A label set must be a brace-wrapped, comma-separated list; a bare
	// `le="+Inf"method="GET"` with no comma is invalid exposition format.
	if !strings.Contains(body, `{le="+Inf",method="GET",path="/health"}`) {
		t.Errorf("expected comma-separated +Inf bucket labels, got:\n%s", body)
	}
}

func TestMetricsHandler_histogramBuckets_areCumulative(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := mp.Meter("test")

	hist, err := meter.Float64Histogram("http_request_duration_seconds")
	if err != nil {
		t.Fatal(err)
	}
	// Default OTel bucket bounds include 5 and 10. One observation lands in
	// (0,5], the other in (5,10]; Prometheus's le="X" bucket must count all
	// observations <= X, i.e. le="10" must include both, not just the one
	// that individually landed in (5,10].
	hist.Record(context.Background(), 1)
	hist.Record(context.Background(), 6)

	handler := telemetry.NewMetricsHandler(reader)
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, `http_request_duration_seconds_bucket{le="10"} 2`) {
		t.Errorf("expected cumulative count 2 for le=10 bucket, got:\n%s", body)
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
