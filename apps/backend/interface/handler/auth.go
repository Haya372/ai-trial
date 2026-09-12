package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

type contextKey string

// ContextKeyUser is the context key used to store the authenticated user.
// Middleware sets this value; handlers read it.
const ContextKeyUser contextKey = "auth_user"

const (
	sessionCookieName = "session_id"
	sessionMaxAge     = 30 * 24 * 60 * 60
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
}

func NewAuthHandler(s SignupExecutor, l LoginExecutor, lo LogoutExecutor) *AuthHandler {
	return &AuthHandler{signup: s, login: l, logout: lo}
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var body api.SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	in := authuc.SignupInput{Email: string(body.Email), Password: body.Password}
	if body.DisplayName != nil {
		in.DisplayName = *body.DisplayName
	}

	out, err := h.signup.Execute(r.Context(), in)
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	setSessionCookie(w, out.SessionID.String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	writeUserResponse(w, out.User)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	out, err := h.login.Execute(r.Context(), authuc.LoginInput{
		Email:    string(body.Email),
		Password: body.Password,
	})
	if err != nil {
		h.writeAuthError(w, err)
		return
	}

	setSessionCookie(w, out.SessionID.String())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeUserResponse(w, out.User)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	sessionID, err := uuid.Parse(cookie.Value)
	if err != nil {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	if err := h.logout.Execute(r.Context(), sessionID); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ContextKeyUser).(user.User)
	if !ok || u == nil {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeUserResponse(w, u)
}

func (h *AuthHandler) writeAuthError(w http.ResponseWriter, err error) {
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
			response.WriteError(w, http.StatusConflict, "CONFLICT", de.Message)
		case user.CodeUserNotFound, user.CodePasswordMismatch:
			response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid email or password")
		default:
			response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		}
	default:
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func writeUserResponse(w http.ResponseWriter, u user.User) {
	if err := json.NewEncoder(w).Encode(api.UserResponse{
		Id:          u.ID(),
		Email:       openapi_types.Email(u.Email()),
		DisplayName: u.DisplayName(),
	}); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}

func setSessionCookie(w http.ResponseWriter, sessionID string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}
