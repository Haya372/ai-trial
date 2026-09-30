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

// resolveSessionUser looks up the user for the session cookie on r. It returns
// (nil, nil) when there is no session or the session is expired/invalid/missing
// its user. It returns a non-nil error only for an unexpected repository failure.
func resolveSessionUser(
	ctx context.Context,
	r *http.Request,
	sessionRepo session.Repository,
	userRepo user.Repository,
	logger *slog.Logger,
) (user.User, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		logger.Warn("session cookie not found", "path", r.URL.Path)
		return nil, nil //nolint:nilerr // no cookie means anonymous, not a failure
	}
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		logger.Warn("invalid session id format in cookie", "path", r.URL.Path)
		return nil, nil //nolint:nilerr // malformed cookie means anonymous, not a failure
	}
	sess, err := sessionRepo.FindActiveByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		logger.Warn("session not found or expired", "path", r.URL.Path)
		return nil, nil
	}
	u, err := userRepo.FindByID(ctx, sess.UserID())
	if err != nil {
		return nil, err
	}
	if u == nil {
		logger.Error("data inconsistency: user not found for active session",
			"user_id", sess.UserID(),
			"path", r.URL.Path,
		)
	}
	return u, nil
}

func RequireAuth(
	sessionRepo session.Repository,
	userRepo user.Repository,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := resolveSessionUser(r.Context(), r, sessionRepo, userRepo, logger)
			if err != nil {
				logger.Error("failed to resolve session user", "error", err, "path", r.URL.Path)
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			if u == nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}
			ctx := context.WithValue(r.Context(), ctxkey.User, u)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(
	sessionRepo session.Repository,
	userRepo user.Repository,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, err := resolveSessionUser(r.Context(), r, sessionRepo, userRepo, logger)
			if err != nil {
				logger.Error("failed to resolve session user in optional auth", "error", err, "path", r.URL.Path)
			}
			if u != nil {
				ctx := context.WithValue(r.Context(), ctxkey.User, u)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}
