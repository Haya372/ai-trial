package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
)

func RequireAuth(
	sessionRepo session.Repository,
	userRepo user.Repository,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_id")
			if err != nil {
				logger.Warn("session cookie not found", "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			sessionID, err := uuid.Parse(cookie.Value)
			if err != nil {
				logger.Warn("invalid session id format in cookie", "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			sess, err := sessionRepo.FindActiveByID(r.Context(), sessionID)
			if err != nil {
				logger.Error("failed to find session", "error", err, "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			if sess == nil {
				logger.Warn("session not found or expired", "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			u, err := userRepo.FindByID(r.Context(), sess.UserID())
			if err != nil {
				logger.Error("failed to find user for session", "error", err, "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			if u == nil {
				logger.Error("data inconsistency: user not found for active session",
					"user_id", sess.UserID(),
					"path", r.URL.Path,
				)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), ctxkey.User, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
