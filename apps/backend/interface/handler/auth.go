package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

const (
	cookieName   = "session_id"
	cookieMaxAge = 30 * 24 * 60 * 60

	errCodeConflict     = "CONFLICT"
	errCodeUnauthorized = "UNAUTHORIZED"
	errCodeInternal     = "INTERNAL_ERROR"
	errCodeValidation   = "VALIDATION_ERROR"
)

type SignupExecutor interface {
	Execute(ctx context.Context, in authuc.SignupInput) (*authuc.AuthOutput, error)
}

type LoginExecutor interface {
	Execute(ctx context.Context, in authuc.LoginInput) (*authuc.AuthOutput, error)
}

type LogoutExecutor interface {
	Execute(ctx context.Context, sessionID uuid.UUID) error
}

type AuthHandler struct {
	signup SignupExecutor
	login  LoginExecutor
	logout LogoutExecutor
	logger *slog.Logger
}

func NewAuthHandler(s SignupExecutor, l LoginExecutor, lo LogoutExecutor, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{signup: s, login: l, logout: lo, logger: logger}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var body api.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	in := authuc.SignupInput{Email: string(body.Email), Password: body.Password}
	if body.DisplayName != nil {
		in.DisplayName = *body.DisplayName
	}

	out, err := h.signup.Execute(r.Context(), in)
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}

	respBody, err := marshalUserResponse(out.User)
	if err != nil {
		h.logger.Error("failed to marshal signup response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	setSessionCookie(w, out.SessionID.String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(respBody)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	out, err := h.login.Execute(r.Context(), authuc.LoginInput{
		Email:    string(body.Email),
		Password: body.Password,
	})
	if err != nil {
		h.writeAuthError(w, r, err)
		return
	}

	respBody, err := marshalUserResponse(out.User)
	if err != nil {
		h.logger.Error("failed to marshal login response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	setSessionCookie(w, out.SessionID.String())
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBody)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		h.logger.Warn("logout attempted without session cookie", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		h.logger.Warn("invalid session id format in logout request", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}
	if err := h.logout.Execute(r.Context(), sessionID); err != nil {
		h.logger.Error("failed to execute logout", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to GET /auth/me", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}
	respBody, err := marshalUserResponse(u)
	if err != nil {
		h.logger.Error("failed to marshal me response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBody) //nolint:gosec
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *domain.ValidationError
	var de *domain.DomainError

	switch {
	case errors.As(err, &ve):
		details := make([]response.ErrorDetail, len(ve.Details))
		for i, d := range ve.Details {
			details[i] = response.ErrorDetail{Field: d.Field, Code: d.Code, Message: d.Message}
		}
		response.WriteValidationError(w, details)
	case errors.As(err, &de):
		switch de.Code {
		case user.CodeEmailTaken:
			response.WriteError(w, http.StatusConflict, errCodeConflict, de.Message)
		case user.CodeUserNotFound, user.CodePasswordMismatch:
			h.logger.Warn("authentication failed", "reason", de.Code, "path", r.URL.Path)
			response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Invalid email or password")
		default:
			h.logger.Error("unexpected domain error in auth handler", "error", err, "path", r.URL.Path)
			response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		}
	default:
		h.logger.Error("internal error in auth handler", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}

func marshalUserResponse(u user.User) ([]byte, error) {
	return json.Marshal(api.UserResponse{
		Id:          u.ID(),
		Email:       openapi_types.Email(u.Email()),
		DisplayName: u.DisplayName(),
	})
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
