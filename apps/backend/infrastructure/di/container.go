package di

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func NewContainer() (*dig.Container, error) {
	c := dig.New()
	for _, p := range []any{newLogger, handler.NewHealthHandler, newRouter} {
		if err := c.Provide(p); err != nil {
			return nil, fmt.Errorf("provide %T: %w", p, err)
		}
	}
	return c, nil
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func newRouter(h *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", h.ServeHTTP)

	return r
}
