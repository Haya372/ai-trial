package auth_test

import "errors"

const (
	testEmail    = "test@example.com"
	testPassword = "SecurePass1!"
)

var (
	errPasswordMismatch = errors.New("mismatch")
	errDBFailure        = errors.New("db error")
)
