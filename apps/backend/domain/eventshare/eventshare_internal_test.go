package eventshare

import (
	"testing"
	"time"
)

// This file uses package eventshare (white-box) instead of eventshare_test
// because isExpired takes an explicit "now" so the exact-equality boundary
// can be pinned deterministically; time.Now() can't do that from outside.

func TestIsExpired_beforeExpiry(t *testing.T) {
	expiresAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := expiresAt.Add(-time.Nanosecond)

	if isExpired(now, expiresAt) {
		t.Error("isExpired() = true, want false one nanosecond before expiresAt")
	}
}

func TestIsExpired_exactlyAtExpiry(t *testing.T) {
	// SPEC-004: expiresAt itself counts as expired (boundary is inclusive).
	expiresAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if !isExpired(expiresAt, expiresAt) {
		t.Error("isExpired() = false, want true when now equals expiresAt")
	}
}

func TestIsExpired_afterExpiry(t *testing.T) {
	expiresAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := expiresAt.Add(time.Nanosecond)

	if !isExpired(now, expiresAt) {
		t.Error("isExpired() = false, want true one nanosecond after expiresAt")
	}
}
