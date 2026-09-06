package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Haya372/ai-trial/backend/infrastructure/di"
)

const (
	defaultPort     = "8080"
	shutdownTimeout = 5 * time.Second
)

func main() {
	c, err := di.NewContainer()
	if err != nil {
		slog.Error("failed to build DI container", "error", err)
		os.Exit(1)
	}

	if err := c.Invoke(run); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func run(r *chi.Mux, logger *slog.Logger) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server failed to start: %w", err)
		}
		return nil
	case <-ctx.Done():
		stop()
	}

	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return <-serverErr
}
