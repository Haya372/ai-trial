package handler_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"
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
