package testutil

import (
	"context"
	"time"
)

// StubTxManager executes the function directly without a real transaction.
// Use in unit tests where actual DB transactions are not needed.
type StubTxManager struct{}

func (s *StubTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// FixedClock is a usecase.Clock that always returns Time.
// Use in unit tests to make time-dependent logic deterministic.
type FixedClock struct {
	Time time.Time
}

func (c FixedClock) Now() time.Time { return c.Time }
