package middleware_test

import (
	"errors"
	"log/slog"
)

var (
	errDB      = errors.New("db error")
	testLogger = slog.New(slog.DiscardHandler)
)
