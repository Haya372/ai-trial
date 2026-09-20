package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

// --- stubs for executor interfaces ---

type stubSignupExec struct {
	fn func(context.Context, authuc.SignupInput) (*authuc.AuthOutput, error)
}

func (s *stubSignupExec) Execute(ctx context.Context, in authuc.SignupInput) (*authuc.AuthOutput, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, in)
}

type stubLoginExec struct {
	fn func(context.Context, authuc.LoginInput) (*authuc.AuthOutput, error)
}

func (s *stubLoginExec) Execute(ctx context.Context, in authuc.LoginInput) (*authuc.AuthOutput, error) {
	if s.fn == nil {
		return nil, nil
	}
	return s.fn(ctx, in)
}

type stubLogoutExec struct {
	fn func(context.Context, uuid.UUID) error
}

func (s *stubLogoutExec) Execute(ctx context.Context, id uuid.UUID) error {
	if s.fn == nil {
		return nil
	}
	return s.fn(ctx, id)
}

// --- stub user.User ---

type stubUser struct {
	id          uuid.UUID
	email       user.Email
	displayName string
}

func (u *stubUser) ID() uuid.UUID                         { return u.id }
func (u *stubUser) Email() user.Email                     { return u.email }
func (u *stubUser) DisplayName() string                   { return u.displayName }
func (u *stubUser) ComparePassword(_ user.Password) error { return nil }

func newStubUser(id uuid.UUID, email, displayName string) user.User {
	return &stubUser{id: id, email: user.Email(email), displayName: displayName}
}

// --- request helpers ---

func signupRequest(t *testing.T, email string) *http.Request {
	t.Helper()
	payload, err := json.Marshal(map[string]string{fieldEmail: email, fieldPassword: testPassword})
	if err != nil {
		t.Fatalf("marshal signup request: %v", err)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/signup", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func loginRequest(t *testing.T, email string) *http.Request {
	t.Helper()
	payload, err := json.Marshal(map[string]string{fieldEmail: email, fieldPassword: testPassword})
	if err != nil {
		t.Fatalf("marshal login request: %v", err)
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/login", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func sessionCookieForRequest(value string) *http.Cookie {
	return &http.Cookie{ //nolint:gosec
		Name:  testSessionCookieName,
		Value: value,
	}
}

// --- Signup tests ---

func TestAuthHandler_Signup_validInput_returns201_and_sets_cookie(t *testing.T) {
	fixedUserID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedSessID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	stub := &stubSignupExec{fn: func(_ context.Context, _ authuc.SignupInput) (*authuc.AuthOutput, error) {
		return &authuc.AuthOutput{
			User:      newStubUser(fixedUserID, "u@ex.com", "U"),
			SessionID: fixedSessID,
		}, nil
	}}
	h := handler.NewAuthHandler(stub, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Signup(rec, signupRequest(t, "u@ex.com"))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	hasCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName && c.Value == fixedSessID.String() {
			hasCookie = true
		}
	}
	if !hasCookie {
		t.Errorf("expected Set-Cookie: %s=%s", testSessionCookieName, fixedSessID)
	}
}

func TestAuthHandler_Signup_validationError_returns400(t *testing.T) {
	stub := &stubSignupExec{fn: func(_ context.Context, _ authuc.SignupInput) (*authuc.AuthOutput, error) {
		return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
			{Field: fieldEmail, Code: "INVALID_FORMAT", Message: "Invalid email format"},
		}}
	}}
	h := handler.NewAuthHandler(stub, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Signup(rec, signupRequest(t, "bad"))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body["code"] != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %v", body["code"])
	}
}

func TestAuthHandler_Signup_emailTaken_returns409(t *testing.T) {
	stub := &stubSignupExec{fn: func(_ context.Context, _ authuc.SignupInput) (*authuc.AuthOutput, error) {
		return nil, user.ErrEmailTaken
	}}
	h := handler.NewAuthHandler(stub, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Signup(rec, signupRequest(t, "dup@ex.com"))

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestAuthHandler_Signup_internalError_returns500(t *testing.T) {
	stub := &stubSignupExec{fn: func(_ context.Context, _ authuc.SignupInput) (*authuc.AuthOutput, error) {
		return nil, errUnexpected
	}}
	h := handler.NewAuthHandler(stub, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Signup(rec, signupRequest(t, "u@ex.com"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestAuthHandler_Signup_responseBody_containsUserFields(t *testing.T) {
	fixedUserID := uuid.MustParse("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")

	stub := &stubSignupExec{fn: func(_ context.Context, _ authuc.SignupInput) (*authuc.AuthOutput, error) {
		return &authuc.AuthOutput{
			User:      newStubUser(fixedUserID, "body@ex.com", "Body"),
			SessionID: uuid.New(),
		}, nil
	}}
	h := handler.NewAuthHandler(stub, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Signup(rec, signupRequest(t, "body@ex.com"))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var body api.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if string(body.Email) != "body@ex.com" {
		t.Errorf("unexpected email: %s", body.Email)
	}
}

// --- Login tests ---

func TestAuthHandler_Login_validCredentials_returns200_and_sets_cookie(t *testing.T) {
	fixedUserID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	fixedSessID := uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc")

	stub := &stubLoginExec{fn: func(_ context.Context, _ authuc.LoginInput) (*authuc.AuthOutput, error) {
		return &authuc.AuthOutput{
			User:      newStubUser(fixedUserID, "u@ex.com", "U"),
			SessionID: fixedSessID,
		}, nil
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, stub, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Login(rec, loginRequest(t, "u@ex.com"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	hasCookie := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName {
			hasCookie = true
		}
	}
	if !hasCookie {
		t.Errorf("expected Set-Cookie: %s", testSessionCookieName)
	}
}

func TestAuthHandler_Login_wrongPassword_returns401(t *testing.T) {
	stub := &stubLoginExec{fn: func(_ context.Context, _ authuc.LoginInput) (*authuc.AuthOutput, error) {
		return nil, user.ErrPasswordMismatch
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, stub, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Login(rec, loginRequest(t, "u@ex.com"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_unknownEmail_returns401(t *testing.T) {
	stub := &stubLoginExec{fn: func(_ context.Context, _ authuc.LoginInput) (*authuc.AuthOutput, error) {
		return nil, user.ErrUserNotFound
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, stub, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Login(rec, loginRequest(t, "no@ex.com"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Login_internalError_returns500(t *testing.T) {
	stub := &stubLoginExec{fn: func(_ context.Context, _ authuc.LoginInput) (*authuc.AuthOutput, error) {
		return nil, errUnexpected
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, stub, &stubLogoutExec{}, testLogger)

	rec := httptest.NewRecorder()
	h.Login(rec, loginRequest(t, "u@ex.com"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- Logout tests ---

func TestAuthHandler_Logout_withValidCookie_returns204_and_clears_cookie(t *testing.T) {
	sessID := uuid.New()
	var deletedID uuid.UUID
	stub := &stubLogoutExec{fn: func(_ context.Context, id uuid.UUID) error {
		deletedID = id
		return nil
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, stub, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
	req.AddCookie(sessionCookieForRequest(sessID.String()))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if deletedID != sessID {
		t.Errorf("expected delete called with %v, got %v", sessID, deletedID)
	}
	cleared := false
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName && (c.MaxAge < 0 || c.Value == "") {
			cleared = true
		}
	}
	if !cleared {
		t.Errorf("expected %s cookie to be cleared", testSessionCookieName)
	}
}

func TestAuthHandler_Logout_noCookie_returns401(t *testing.T) {
	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Logout_invalidCookieUUID_returns401(t *testing.T) {
	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
	req.AddCookie(sessionCookieForRequest("not-a-uuid"))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Logout_executionError_returns500(t *testing.T) {
	stub := &stubLogoutExec{fn: func(_ context.Context, _ uuid.UUID) error {
		return errInternal
	}}
	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, stub, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
	req.AddCookie(sessionCookieForRequest(uuid.New().String()))
	rec := httptest.NewRecorder()
	h.Logout(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

// --- GetMe tests ---

func TestAuthHandler_GetMe_withUserInContext_returns200_and_user_body(t *testing.T) {
	fixedUserID := uuid.MustParse("dddddddd-dddd-dddd-dddd-dddddddddddd")
	u := newStubUser(fixedUserID, "me@ex.com", "Me")

	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, u))
	rec := httptest.NewRecorder()
	h.GetMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body api.UserResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if string(body.Email) != "me@ex.com" {
		t.Errorf("expected email me@ex.com, got %s", body.Email)
	}
	if body.DisplayName != "Me" {
		t.Errorf("expected display name Me, got %s", body.DisplayName)
	}
}

func TestAuthHandler_GetMe_withoutUser_returns401(t *testing.T) {
	h := handler.NewAuthHandler(&stubSignupExec{}, &stubLoginExec{}, &stubLogoutExec{}, testLogger)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	h.GetMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
