package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/dig"

	"github.com/Haya372/ai-trial/backend/interface/handler"
)

func main() {
	c := dig.New()

	c.Provide(newLogger)
	c.Provide(handler.NewHealthHandler)
	c.Provide(newRouter)

	if err := c.Invoke(run); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
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

func run(r *chi.Mux, logger *slog.Logger) error {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
