package handler_test

import "errors"

const (
	testSessionCookieName = "session_id"
	testPassword          = "SecurePass1!"
	fieldEmail            = "email"
	fieldPassword         = "password"
)

var errUnexpected = errors.New("unexpected error")
