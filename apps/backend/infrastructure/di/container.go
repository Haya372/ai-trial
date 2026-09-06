package di

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func NewContainer(ctx context.Context) (*dig.Container, error) {
	c := dig.New()
	for _, p := range []any{
		func() context.Context { return ctx },
		newLogger, newPool, handler.NewHealthHandler, newRouter,
	} {
		if err := c.Provide(p); err != nil {
			return nil, fmt.Errorf("provide %T: %w", p, err)
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

func newRouter(h *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", h.ServeHTTP)

	return r
}
