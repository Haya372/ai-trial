package eventshare_test

import (
	"errors"
	"log/slog"
)

var (
	errDBFailure = errors.New("db error")
	testLogger   = slog.New(slog.DiscardHandler)
)
