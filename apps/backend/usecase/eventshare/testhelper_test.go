package eventshare_test

import (
	"context"
	"errors"
	"log/slog"
)

var (
	errDBFailure = errors.New("db error")
	testLogger   = slog.New(slog.DiscardHandler)
)

// failingTxManager returns errDBFailure without ever invoking fn, simulating
// a transaction that fails to start.
type failingTxManager struct{}

func (f *failingTxManager) RunInTx(context.Context, func(context.Context) error) error {
	return errDBFailure
}
