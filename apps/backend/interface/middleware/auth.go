package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
)

func RequireAuth(sessionRepo session.Repository, userRepo user.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			sess, err := sessionRepo.FindActiveByID(r.Context(), sessionID)
			if err != nil || sess == nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			u, err := userRepo.FindByID(r.Context(), sess.UserID())
			if err != nil || u == nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), handler.ContextKeyUser, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
