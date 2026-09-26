package middleware

import (
	"net/http"
	"strconv"

	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracing returns a middleware that starts an "http.request" span for every
// request using the given TracerProvider, tagging it with the chi route
// pattern (not the raw path, to keep span/metric cardinality bounded).
func Tracing(tp trace.TracerProvider) func(http.Handler) http.Handler {
	tracer := tp.Tracer("github.com/Haya372/ai-trial/backend")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "http.request")
			defer span.End()

			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			req := r.WithContext(ctx)
			next.ServeHTTP(ww, req)

			status := statusOf(ww)

			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", routePatternOf(req)),
				attribute.Int("http.status_code", status),
			)
			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, strconv.Itoa(status))
			}
		})
	}
}
