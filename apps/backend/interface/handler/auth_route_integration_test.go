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

func signupAndGetCookie(t *testing.T, router *chi.Mux, email string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": testPassword})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
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

func TestRoute_Logout_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRoute_Signup_then_Logout_succeeds(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	cookie := signupAndGetCookie(t, router, "logout@ex.com")

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_GetMe_withoutSession_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildRouteTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
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
