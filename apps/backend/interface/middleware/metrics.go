package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Metrics returns a middleware that records RED (Rate, Errors, Duration)
// metrics for every HTTP request using the given MeterProvider.
//
// Label cardinality is kept safe by using chi's route pattern (e.g.
// /events/{id}) instead of the raw request path.
func Metrics(mp metric.MeterProvider) func(http.Handler) http.Handler {
	meter := mp.Meter("github.com/Haya372/ai-trial/backend")

	requestsTotal, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	errorsTotal, _ := meter.Int64Counter(
		"http_errors_total",
		metric.WithDescription("Total number of HTTP 4xx/5xx responses"),
	)
	durationSecs, _ := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request latency in seconds"),
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}
			elapsed := time.Since(start).Seconds()

			// Use chi's route pattern to avoid cardinality explosion
			// from path parameters such as /events/abc-123.
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = r.URL.Path
			}

			attrs := []attribute.KeyValue{
				attribute.String("method", r.Method),
				attribute.String("path", routePattern),
				attribute.String("status", strconv.Itoa(status)),
			}
			attrOpt := metric.WithAttributes(attrs...)

			requestsTotal.Add(r.Context(), 1, attrOpt)
			if status >= 400 {
				errorsTotal.Add(r.Context(), 1, attrOpt)
			}
			durationSecs.Record(r.Context(), elapsed,
				metric.WithAttributes(
					attribute.String("method", r.Method),
					attribute.String("path", routePattern),
				),
			)
		})
	}
}
