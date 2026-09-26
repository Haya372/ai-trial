package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

type tracingFixture struct {
	router   *chi.Mux
	exporter *tracetest.InMemoryExporter
}

func newTracingFixture(t *testing.T, route string, status int) tracingFixture {
	t.Helper()
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	t.Cleanup(func() { _ = tp.Shutdown(t.Context()) })

	r := chi.NewRouter()
	r.Use(middleware.Tracing(tp))
	r.Get(route, handlerWithStatus(status).ServeHTTP)

	return tracingFixture{router: r, exporter: exp}
}

func TestTracingMW_recordsSpanNamedHTTPRequest(t *testing.T) {
	fx := newTracingFixture(t, "/events", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	spans := fx.exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "http.request" {
		t.Errorf("expected span name http.request, got %s", spans[0].Name)
	}
}

func TestTracingMW_setsRouteAttributes(t *testing.T) {
	fx := newTracingFixture(t, "/events/{id}", http.StatusOK)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events/abc-123", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	spans := fx.exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	attrs := map[string]string{}
	for _, kv := range spans[0].Attributes {
		attrs[string(kv.Key)] = kv.Value.String()
	}

	if attrs["http.method"] != "GET" {
		t.Errorf("expected http.method=GET, got %q", attrs["http.method"])
	}
	if attrs["http.route"] != "/events/{id}" {
		t.Errorf("expected http.route=/events/{id}, got %q", attrs["http.route"])
	}
	if attrs["http.status_code"] != "200" {
		t.Errorf("expected http.status_code=200, got %q", attrs["http.status_code"])
	}
}

func TestTracingMW_setsErrorStatus_on5xx(t *testing.T) {
	fx := newTracingFixture(t, "/events", http.StatusInternalServerError)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/events", nil)
	fx.router.ServeHTTP(httptest.NewRecorder(), req)

	spans := fx.exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status.Code != codes.Error {
		t.Errorf("expected span status Error, got %v", spans[0].Status.Code)
	}
}
