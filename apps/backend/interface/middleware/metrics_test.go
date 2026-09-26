package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

// metricsFixture creates a ManualReader-backed MeterProvider and a chi
// router with the Metrics middleware pre-registered on the given route.
type metricsFixture struct {
	router *chi.Mux
	reader *sdkmetric.ManualReader
}

func newMetricsFixture(t *testing.T, route string, status int) metricsFixture {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { _ = mp.Shutdown(t.Context()) })

	r := chi.NewRouter()
	r.Use(middleware.Metrics(mp))
	r.Get(route, handlerWithStatus(status).ServeHTTP)

	return metricsFixture{router: r, reader: reader}
}

func (f metricsFixture) collect(t *testing.T) metricdata.ResourceMetrics {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := f.reader.Collect(t.Context(), &rm); err != nil {
		t.Fatalf("collect metrics: %v", err)
	}
	return rm
}

func findMetric(rm *metricdata.ResourceMetrics, name string) (metricdata.Metrics, bool) {
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m, true
			}
		}
	}
	return metricdata.Metrics{}, false
}

func sumDataPoints(m metricdata.Metrics) int64 {
	if sum, ok := m.Data.(metricdata.Sum[int64]); ok {
		var total int64
		for _, dp := range sum.DataPoints {
			total += dp.Value
		}
		return total
	}
	return 0
}

func TestMetricsMW_countsRequest(t *testing.T) {
	fx := newMetricsFixture(t, "/events", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	rm := fx.collect(t)
	m, ok := findMetric(&rm, "http_requests_total")
	if !ok {
		t.Fatal("http_requests_total metric not found")
	}
	if got := sumDataPoints(m); got != 1 {
		t.Errorf("http_requests_total: expected 1, got %d", got)
	}
}

func TestMetricsMW_countsError_on4xx(t *testing.T) {
	fx := newMetricsFixture(t, "/events", http.StatusBadRequest)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	rm := fx.collect(t)
	m, ok := findMetric(&rm, "http_errors_total")
	if !ok {
		t.Fatal("http_errors_total metric not found")
	}
	if got := sumDataPoints(m); got != 1 {
		t.Errorf("http_errors_total: expected 1, got %d", got)
	}
}

func TestMetricsMW_doesNotCountError_on2xx(t *testing.T) {
	fx := newMetricsFixture(t, "/events", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	rm := fx.collect(t)
	if _, ok := findMetric(&rm, "http_errors_total"); ok {
		t.Error("http_errors_total should not have data points for 2xx response")
	}
}

func TestMetricsMW_recordsDuration(t *testing.T) {
	fx := newMetricsFixture(t, "/events", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	rm := fx.collect(t)
	m, ok := findMetric(&rm, "http_request_duration_seconds")
	if !ok {
		t.Fatal("http_request_duration_seconds metric not found")
	}
	hist, ok := m.Data.(metricdata.Histogram[float64])
	if !ok {
		t.Fatal("http_request_duration_seconds is not a Histogram")
	}
	if len(hist.DataPoints) == 0 || hist.DataPoints[0].Count == 0 {
		t.Error("http_request_duration_seconds should have recorded a data point")
	}
}

func TestMetricsMW_usesRoutePattern_notRawPath(t *testing.T) {
	fx := newMetricsFixture(t, "/events/{id}", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events/abc-123", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	rm := fx.collect(t)
	m, ok := findMetric(&rm, "http_requests_total")
	if !ok {
		t.Fatal("http_requests_total metric not found")
	}
	sum, ok := m.Data.(metricdata.Sum[int64])
	if !ok || len(sum.DataPoints) == 0 {
		t.Fatal("no data points in http_requests_total")
	}

	dp := sum.DataPoints[0]
	pathVal, ok2 := (&dp.Attributes).Value("path")
	if !ok2 {
		t.Fatal("path attribute not found in data points")
	}
	if pathVal.AsString() != "/events/{id}" {
		t.Errorf("expected path=/events/{id}, got path=%s", pathVal.AsString())
	}
}
