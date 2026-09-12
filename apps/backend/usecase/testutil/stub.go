package testutil

import "context"

// StubTxManager executes the function directly without a real transaction.
// Use in unit tests where actual DB transactions are not needed.
type StubTxManager struct{}

func (s *StubTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
