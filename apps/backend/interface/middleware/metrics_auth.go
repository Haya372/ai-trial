package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/Haya372/ai-trial/backend/interface/handler/response"
)

const bearerPrefix = "Bearer "

// RequireMetricsToken protects an endpoint with a static Bearer token.
// An empty token rejects every request so a missing configuration never
// leaves the endpoint open.
func RequireMetricsToken(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !hasValidBearer(r.Header.Get("Authorization"), token) {
				w.Header().Set("WWW-Authenticate", "Bearer")
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasValidBearer(header, token string) bool {
	if token == "" || !strings.HasPrefix(header, bearerPrefix) {
		return false
	}
	got := strings.TrimPrefix(header, bearerPrefix)
	return subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
}
