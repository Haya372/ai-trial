package auth_test

import "errors"

const (
	testEmail    = "test@example.com"
	testPassword = "SecurePass1!"
)

var errDBFailure = errors.New("db error")
