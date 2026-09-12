package auth_test

import (
	"context"
	"errors"
)

const (
	testEmail    = "test@example.com"
	testPassword = "SecurePass1!"
)

var (
	errPasswordMismatch = errors.New("mismatch")
	errDBFailure        = errors.New("db error")
)

type stubTxManager struct{}

func (s *stubTxManager) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}
