// Package route exposes the matched chi route of a request for logging,
// metrics and tracing.
package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Unmatched is reported for requests chi routed but matched no route (404/405).
const Unmatched = "unmatched"

// Pattern returns chi's matched route pattern (e.g. "/events/{id}"). Using
// the pattern instead of the raw path keeps metric/span label cardinality
// bounded (see observability-guidelines.md) and keeps secrets carried in the
// path, such as share tokens in "/shares/{token}", out of logs.
//
// A request chi routed without a match (404/405) yields Unmatched rather than
// the raw path, which may still carry a token (e.g. DELETE /shares/{token}).
// Only a request outside any chi router falls back to the raw path.
func Pattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx == nil {
		return r.URL.Path
	}
	if p := rctx.RoutePattern(); p != "" {
		return p
	}
	return Unmatched
}
