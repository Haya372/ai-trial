package di

import (
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func NewContainer() *dig.Container {
	c := dig.New()
	c.Provide(newLogger)
	c.Provide(handler.NewHealthHandler)
	c.Provide(newRouter)
	return c
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func newRouter(h *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", h.ServeHTTP)
	return r
}
