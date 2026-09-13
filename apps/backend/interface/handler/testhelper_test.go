package handler_test

import (
	"errors"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	testSessionCookieName = "session_id"
	testPassword          = "SecurePass1!"
	fieldEmail            = "email"
	fieldPassword         = "password"
)

var (
	errUnexpected = errors.New("unexpected error")
	errInternal   = errors.New("internal error")
)

func mustParseDate(t *testing.T, s string) openapi_types.Date {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return openapi_types.Date{Time: parsed}
}
