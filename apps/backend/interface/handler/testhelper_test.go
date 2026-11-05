package handler_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	testSessionCookieName = "session_id"
	testPassword          = "SecurePass1!"
	fieldEmail            = "email"
	fieldPassword         = "password"
	codeValidationError   = "VALIDATION_ERROR"
	codeInvalidDateRange  = "INVALID_DATE_RANGE"
)

var (
	errUnexpected = errors.New("unexpected error")
	errInternal   = errors.New("internal error")
	testLogger    = slog.New(slog.DiscardHandler)
)

func mustParseDate(t *testing.T, s string) openapi_types.Date {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return openapi_types.Date{Time: parsed}
}

// secretShareToken is a share token value that must never appear in logs.
const secretShareToken = "s3cr3t-share-token-value"

// newBufferLogger returns a logger that writes JSON records into the returned
// buffer, so tests can assert on everything a handler logged.
func newBufferLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})), &buf
}

func assertShareTokenNotLogged(t *testing.T, logs *bytes.Buffer) {
	t.Helper()
	if logs.Len() == 0 {
		t.Fatal("expected the handler to log something")
	}
	if strings.Contains(logs.String(), secretShareToken) {
		t.Fatalf("share token leaked into logs:\n%s", logs.String())
	}
}
