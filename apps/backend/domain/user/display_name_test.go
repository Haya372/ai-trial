package user_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestNewDisplayName_valid(t *testing.T) {
	name, err := user.NewDisplayName("hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if name != "hello" {
		t.Errorf("got %q, want %q", name, "hello")
	}
}

func TestNewDisplayName_empty(t *testing.T) {
	_, err := user.NewDisplayName("")
	if err != nil {
		t.Errorf("empty display name should be valid, got: %v", err)
	}
}

func TestNewDisplayName_maxLength(t *testing.T) {
	_, err := user.NewDisplayName(strings.Repeat("a", 50))
	if err != nil {
		t.Errorf("50-char display name should be valid, got: %v", err)
	}
}

func TestNewDisplayName_tooLong(t *testing.T) {
	_, err := user.NewDisplayName(strings.Repeat("a", 51))
	if !errors.Is(err, user.ErrDisplayNameTooLong) {
		t.Errorf("got %v, want ErrDisplayNameTooLong", err)
	}
}
