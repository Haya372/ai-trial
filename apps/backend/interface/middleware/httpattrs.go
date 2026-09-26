package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

// statusOf returns the wrapped response's status code, defaulting to 200 OK
// when the handler never explicitly called WriteHeader.
func statusOf(ww chimw.WrapResponseWriter) int {
	if status := ww.Status(); status != 0 {
		return status
	}
	return http.StatusOK
}

// routePatternOf returns chi's matched route pattern (e.g. "/events/{id}"),
// falling back to the raw request path when chi has no route context. Using
// the pattern instead of the raw path keeps metric/span label cardinality
// bounded (see observability-guidelines.md).
func routePatternOf(r *http.Request) string {
	if p := chi.RouteContext(r.Context()).RoutePattern(); p != "" {
		return p
	}
	return r.URL.Path
}
