package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	"github.com/Haya372/ai-trial/backend/interface/middleware"
)

// --- stub session.Repository ---

type stubSessionRepo struct {
	findActiveFn func(ctx context.Context, id uuid.UUID) (session.Session, error)
}

func (r *stubSessionRepo) Create(_ context.Context, _ uuid.UUID, _ time.Time) (session.Session, error) {
	return nil, nil
}
func (r *stubSessionRepo) FindByID(_ context.Context, _ uuid.UUID) (session.Session, error) {
	return nil, nil
}
func (r *stubSessionRepo) FindActiveByID(ctx context.Context, id uuid.UUID) (session.Session, error) {
	if r.findActiveFn == nil {
		return nil, nil
	}
	return r.findActiveFn(ctx, id)
}
func (r *stubSessionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }

// --- stub session.Session ---

type stubSession struct {
	id        uuid.UUID
	userID    uuid.UUID
	expiresAt time.Time
}

func (s *stubSession) ID() uuid.UUID        { return s.id }
func (s *stubSession) UserID() uuid.UUID    { return s.userID }
func (s *stubSession) ExpiresAt() time.Time { return s.expiresAt }
func (s *stubSession) IsExpired() bool      { return time.Now().After(s.expiresAt) }

// --- stub user.Repository ---

type stubUserRepo struct {
	findByIDFn func(ctx context.Context, id uuid.UUID) (user.User, error)
}

func (r *stubUserRepo) Create(_ context.Context, _ user.Email, _ string, _ user.Password) (user.User, error) {
	return nil, nil
}
func (r *stubUserRepo) FindByEmail(_ context.Context, _ user.Email) (user.User, error) {
	return nil, nil
}
func (r *stubUserRepo) FindByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	if r.findByIDFn == nil {
		return nil, nil
	}
	return r.findByIDFn(ctx, id)
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

// --- helpers ---

func nextHandlerCapture(called *bool, capturedUser *user.User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		u, _ := r.Context().Value(handler.ContextKeyUser).(user.User)
		*capturedUser = u
		w.WriteHeader(http.StatusOK)
	})
}

const testSessionCookieName = "session_id"

func requestWithCookie(t *testing.T, cookieValue string) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/auth/me", nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: testSessionCookieName, Value: cookieValue}) //nolint:gosec
	}
	return req
}

// --- Tests ---

func TestRequireAuth_validSession_setsUserInContext_and_calls_next(t *testing.T) {
	sessID := uuid.New()
	userID := uuid.New()

	sess := &stubSession{id: sessID, userID: userID, expiresAt: time.Now().Add(time.Hour)}
	u := &stubUser{id: userID, email: "u@ex.com", displayName: "U"}

	sessRepo := &stubSessionRepo{findActiveFn: func(_ context.Context, id uuid.UUID) (session.Session, error) {
		if id == sessID {
			return sess, nil
		}
		return nil, nil
	}}
	userRepo := &stubUserRepo{findByIDFn: func(_ context.Context, id uuid.UUID) (user.User, error) {
		if id == userID {
			return u, nil
		}
		return nil, nil
	}}

	called := false
	var capturedUser user.User
	mw := middleware.RequireAuth(sessRepo, userRepo)
	handler := mw(nextHandlerCapture(&called, &capturedUser))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if capturedUser == nil || capturedUser.ID() != userID {
		t.Errorf("expected user %v in context, got %v", userID, capturedUser)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRequireAuth_noSessionCookie_returns401(t *testing.T) {
	mw := middleware.RequireAuth(&stubSessionRepo{}, &stubUserRepo{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, "")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_invalidUUID_returns401(t *testing.T) {
	mw := middleware.RequireAuth(&stubSessionRepo{}, &stubUserRepo{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, "not-a-uuid")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_sessionNotFound_returns401(t *testing.T) {
	sessRepo := &stubSessionRepo{findActiveFn: func(_ context.Context, _ uuid.UUID) (session.Session, error) {
		return nil, nil // not found
	}}
	mw := middleware.RequireAuth(sessRepo, &stubUserRepo{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_sessionRepoError_returns401(t *testing.T) {
	sessRepo := &stubSessionRepo{findActiveFn: func(_ context.Context, _ uuid.UUID) (session.Session, error) {
		return nil, errDB
	}}
	mw := middleware.RequireAuth(sessRepo, &stubUserRepo{})
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, uuid.New().String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_userNotFound_returns401(t *testing.T) {
	sessID := uuid.New()
	userID := uuid.New()
	sess := &stubSession{id: sessID, userID: userID, expiresAt: time.Now().Add(time.Hour)}

	sessRepo := &stubSessionRepo{findActiveFn: func(_ context.Context, _ uuid.UUID) (session.Session, error) {
		return sess, nil
	}}
	userRepo := &stubUserRepo{findByIDFn: func(_ context.Context, _ uuid.UUID) (user.User, error) {
		return nil, user.ErrUserNotFound
	}}

	mw := middleware.RequireAuth(sessRepo, userRepo)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("next handler should not be called")
	}))

	req := requestWithCookie(t, sessID.String())
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
