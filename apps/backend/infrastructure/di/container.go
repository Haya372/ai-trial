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
	mustProvide(c, newLogger)
	mustProvide(c, handler.NewHealthHandler)
	mustProvide(c, newRouter)
	return c
}

func mustProvide(c *dig.Container, constructor any) {
	if err := c.Provide(constructor); err != nil {
		panic(err)
	}
}

func newLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func newRouter(h *handler.HealthHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/health", h.ServeHTTP)
	return r
}
