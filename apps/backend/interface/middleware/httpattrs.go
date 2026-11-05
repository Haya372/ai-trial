package middleware

import (
	"net/http"

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
