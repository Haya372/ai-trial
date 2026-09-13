//go:build integration

package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
)

var routeTestPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	c, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic("start postgres container: " + err.Error())
	}
	defer func() { _ = c.Terminate(ctx) }()

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("get connection string: " + err.Error())
	}

	routeTestPool, err = db.NewPool(ctx, dsn)
	if err != nil {
		panic("connect to postgres: " + err.Error())
	}
	defer routeTestPool.Close()

	migrateDSN := strings.Replace(dsn, "postgres://", "pgx5://", 1)
	mg, err := migrate.New("file://../../db/migrations", migrateDSN)
	if err != nil {
		panic("create migrate: " + err.Error())
	}
	if err := mg.Up(); err != nil && err != migrate.ErrNoChange {
		panic("migrate up: " + err.Error())
	}

	os.Exit(m.Run())
}

func buildRouteTestRouter() *chi.Mux {
	userRepo := repository.NewUserRepository(routeTestPool)
	sessRepo := repository.NewSessionRepository(routeTestPool)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr)
	login := authuc.NewLoginCommand(userRepo, sessRepo)
	logout := authuc.NewLogoutCommand(sessRepo)

	auth := handler.NewAuthHandler(signup, login, logout)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Post("/auth/logout", auth.Logout)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/auth/me", auth.GetMe)
	return r
}

func setupRouteTest(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_, err := routeTestPool.Exec(context.Background(), "TRUNCATE TABLE sessions, users RESTART IDENTITY CASCADE")
		if err != nil {
			t.Errorf("truncate tables: %v", err)
		}
	})
}

// --- DB assertion helpers ---

func assertSessionExistsInDB(t *testing.T, sessionID string) {
	t.Helper()
	var count int
	err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM sessions WHERE id = $1", sessionID).Scan(&count)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if count == 0 {
		t.Errorf("expected session %s to exist in DB", sessionID)
	}
}

func assertSessionNotExistsInDB(t *testing.T, sessionID string) {
	t.Helper()
	var count int
	err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM sessions WHERE id = $1", sessionID).Scan(&count)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}
	if count != 0 {
		t.Errorf("expected session %s to be deleted from DB, but it still exists", sessionID)
	}
}

func assertUserExistsInDB(t *testing.T, email string) {
	t.Helper()
	var count int
	err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		t.Fatalf("query user: %v", err)
	}
	if count == 0 {
		t.Errorf("expected user with email %s to exist in DB", email)
	}
}

// --- Request / response helpers ---

func authBody(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequestWithContext(t.Context(), method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func signupAndGetCookie(t *testing.T, router *chi.Mux, email string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, signupRequest(t, email))
	if rec.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName {
			return c
		}
	}
	t.Fatal("no session cookie in signup response")
	return nil
}

func responseCode(t *testing.T, router *chi.Mux, req *http.Request) int {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

// --- Signup ---

func TestRoute_Signup_validInput_creates_user_and_session_in_DB(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, signupRequest(t, "new@ex.com"))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	assertUserExistsInDB(t, "new@ex.com")

	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie")
	}
	assertSessionExistsInDB(t, sessionCookie.Value)
}

func TestRoute_Signup_errorCases(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	// pre-create user for duplicate test
	router.ServeHTTP(httptest.NewRecorder(), signupRequest(t, "dup@ex.com"))

	tests := []struct {
		name        string
		reqFn       func() *http.Request
		wantCode    int
		wantErrCode string
	}{
		{
			name: "invalid email returns 400 with validation error",
			reqFn: func() *http.Request {
				return authBody(t, http.MethodPost, "/auth/signup", map[string]string{
					"email": "not-an-email", "password": testPassword,
				})
			},
			wantCode:    http.StatusBadRequest,
			wantErrCode: "VALIDATION_ERROR",
		},
		{
			name: "weak password returns 400 with validation error",
			reqFn: func() *http.Request {
				return authBody(t, http.MethodPost, "/auth/signup", map[string]string{
					"email": "weak@ex.com", "password": "password",
				})
			},
			wantCode:    http.StatusBadRequest,
			wantErrCode: "VALIDATION_ERROR",
		},
		{
			name:     "duplicate email returns 409",
			reqFn:    func() *http.Request { return signupRequest(t, "dup@ex.com") },
			wantCode: http.StatusConflict,
		},
		{
			name: "invalid JSON returns 400",
			reqFn: func() *http.Request {
				req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/signup",
					strings.NewReader("not json"))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, tt.reqFn())
			if rec.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, rec.Code, rec.Body.String())
			}
			if tt.wantErrCode != "" {
				var body map[string]any
				_ = json.NewDecoder(rec.Body).Decode(&body)
				if body["code"] != tt.wantErrCode {
					t.Errorf("expected %s, got %v", tt.wantErrCode, body["code"])
				}
			}
		})
	}
}

// --- Login ---

func TestRoute_Login_validCredentials_returns200_and_new_session_in_DB(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	signupAndGetCookie(t, router, "login@ex.com")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, loginRequest(t, "login@ex.com"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == testSessionCookieName {
			assertSessionExistsInDB(t, c.Value)
			return
		}
	}
	t.Error("no session cookie in login response")
}

func TestRoute_Login_errorCases(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	// pre-create user for wrong password test
	signupAndGetCookie(t, router, "wp@ex.com")

	tests := []struct {
		name     string
		reqFn    func() *http.Request
		wantCode int
	}{
		{
			name: "wrong password returns 401",
			reqFn: func() *http.Request {
				return authBody(t, http.MethodPost, "/auth/login", map[string]string{
					"email": "wp@ex.com", "password": "WrongPass1!",
				})
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "unknown email returns 401",
			reqFn:    func() *http.Request { return loginRequest(t, "nobody@ex.com") },
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "invalid JSON returns 400",
			reqFn: func() *http.Request {
				req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/login",
					strings.NewReader("not json"))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if code := responseCode(t, router, tt.reqFn()); code != tt.wantCode {
				t.Fatalf("expected %d, got %d", tt.wantCode, code)
			}
		})
	}
}

// --- Logout ---

func TestRoute_Logout_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	if code := responseCode(t, router, httptest.NewRequest(http.MethodPost, "/auth/logout", nil)); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_Logout_deletesSessionFromDB(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	cookie := signupAndGetCookie(t, router, "dellogout@ex.com")
	assertSessionExistsInDB(t, cookie.Value)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	assertSessionNotExistsInDB(t, cookie.Value)
}

func TestRoute_Logout_reusingSessionAfterLogout_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	cookie := signupAndGetCookie(t, router, "reuse@ex.com")

	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.AddCookie(cookie)
	router.ServeHTTP(httptest.NewRecorder(), logoutReq)

	getMeReq := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	getMeReq.AddCookie(cookie)
	if code := responseCode(t, router, getMeReq); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", code)
	}
}

// --- GetMe ---

func TestRoute_GetMe_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	if code := responseCode(t, router, httptest.NewRequest(http.MethodGet, "/auth/me", nil)); code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestRoute_Signup_then_GetMe_returns_user(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	cookie := signupAndGetCookie(t, router, "getme@ex.com")

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["email"] != "getme@ex.com" {
		t.Errorf("expected email getme@ex.com, got %v", body["email"])
	}
}
