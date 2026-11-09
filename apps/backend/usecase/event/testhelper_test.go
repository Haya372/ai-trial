package event_test

import (
	"context"
	"errors"
	"log/slog"
)

var (
	errDBFailure = errors.New("db error")
	testLogger   = slog.New(slog.DiscardHandler)
)

const (
	testLocation = "Tokyo"
	testURL      = "https://example.com"
)

type inTxKey struct{}

// markingTxManager marks the context passed to fn, so a repository mock can
// assert that it was called inside RunInTx.
type markingTxManager struct{}

func (markingTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(context.WithValue(ctx, inTxKey{}, true))
}

func inTx(ctx context.Context) bool {
	marked, _ := ctx.Value(inTxKey{}).(bool)
	return marked
}
