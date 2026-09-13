package di

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/domain/session"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	"github.com/Haya372/ai-trial/backend/usecase"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func NewContainer(ctx context.Context) (*dig.Container, error) {
	c := dig.New()
	for _, p := range []any{
		func() context.Context { return ctx },
		newLogger,
		newPool,
		newTxManager,
		repository.NewUserRepository,
		repository.NewSessionRepository,
		repository.NewEventQueryRepository,
		newSignupExecutor,
		newLoginExecutor,
		newLogoutExecutor,
		newListEventsExecutor,
		handler.NewHealthHandler,
		handler.NewAuthHandler,
		handler.NewEventHandler,
		newRouter,
	} {
		if err := c.Provide(p); err != nil {
			return nil, fmt.Errorf("provide: %w", err)
		}
	}
	return c, nil
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func newPool(ctx context.Context, logger *slog.Logger) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	logger.Info("connected to database")
	return pool, nil
}

func newTxManager(pool *pgxpool.Pool) usecase.TransactionManager {
	return db.NewPgxTxManager(pool)
}

func newSignupExecutor(
	ur user.Repository,
	sr session.Repository,
	tx usecase.TransactionManager,
) handler.SignupExecutor {
	return authuc.NewSignupCommand(ur, sr, tx)
}

func newLoginExecutor(ur user.Repository, sr session.Repository) handler.LoginExecutor {
	return authuc.NewLoginCommand(ur, sr)
}

func newLogoutExecutor(sr session.Repository) handler.LogoutExecutor {
	return authuc.NewLogoutCommand(sr)
}

func newListEventsExecutor(s eventuc.QueryService) handler.ListEventsExecutor {
	return eventuc.NewListEventsQuery(s)
}

func newRouter(
	health *handler.HealthHandler,
	auth *handler.AuthHandler,
	ev *handler.EventHandler,
	sessRepo session.Repository,
	userRepo user.Repository,
) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", health.ServeHTTP)
	r.Post("/auth/signup", auth.Signup)
	r.Post("/auth/login", auth.Login)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Post("/auth/logout", auth.Logout)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo)).Get("/events", ev.ServeHTTP)
	return r
}
