package user_test

import (
	"strings"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestNewPassword_valid(t *testing.T) {
	p, err := user.NewPassword("SecurePass1!")
	if err != nil {
		t.Fatalf("NewPassword() unexpected error: %v", err)
	}
	if p.Hash() == "" {
		t.Error("Hash() must not be empty")
	}
	if p.Hash() == "SecurePass1!" {
		t.Error("Hash() must not equal plaintext")
	}
}

func TestNewPassword_tooShort(t *testing.T) {
	_, err := user.NewPassword("short")
	if err == nil {
		t.Error("NewPassword() expected error for too-short password")
	}
}

func TestNewPassword_tooLong(t *testing.T) {
	_, err := user.NewPassword(strings.Repeat("a", 129))
	if err == nil {
		t.Error("NewPassword() expected error for too-long password")
	}
}

func TestNewPasswordFromHash_returnsHash(t *testing.T) {
	p := user.NewPasswordFromHash("$2a$10$somehashvalue")
	if p.Hash() != "$2a$10$somehashvalue" {
		t.Errorf("Hash() = %q, want %q", p.Hash(), "$2a$10$somehashvalue")
	}
}
