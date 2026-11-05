package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Haya372/ai-trial/backend/interface/route"
)

func patternVia(t *testing.T, method, path string) string {
	t.Helper()
	var got string
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req)
			got = route.Pattern(req)
		})
	})
	r.Get("/shares/{token}", func(http.ResponseWriter, *http.Request) {})
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), method, path, nil))
	return got
}

func TestPattern_matchedRoute_returnsPattern(t *testing.T) {
	if got := patternVia(t, http.MethodGet, "/shares/secret"); got != "/shares/{token}" {
		t.Errorf("expected /shares/{token}, got %q", got)
	}
}

func TestPattern_methodNotAllowed_returnsUnmatched(t *testing.T) {
	if got := patternVia(t, http.MethodDelete, "/shares/secret"); got != route.Unmatched {
		t.Errorf("expected %q, got %q", route.Unmatched, got)
	}
}

func TestPattern_notFound_returnsUnmatched(t *testing.T) {
	if got := patternVia(t, http.MethodGet, "/unknown/secret"); got != route.Unmatched {
		t.Errorf("expected %q, got %q", route.Unmatched, got)
	}
}

func TestPattern_noRouteContext_returnsRawPath(t *testing.T) {
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	if got := route.Pattern(req); got != "/health" {
		t.Errorf("expected /health, got %q", got)
	}
}
