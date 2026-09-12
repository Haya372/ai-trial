package user_test

import (
	"errors"
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
	if !errors.Is(err, user.ErrPasswordTooShort) {
		t.Errorf("NewPassword() error = %v, want ErrPasswordTooShort", err)
	}
}

func TestNewPassword_tooLong(t *testing.T) {
	_, err := user.NewPassword(strings.Repeat("a", 73))
	if !errors.Is(err, user.ErrPasswordTooLong) {
		t.Errorf("NewPassword() error = %v, want ErrPasswordTooLong", err)
	}
}

func TestNewPassword_maxLength(t *testing.T) {
	_, err := user.NewPassword(strings.Repeat("a", 72))
	if err != nil {
		t.Errorf("NewPassword() unexpected error for 72-char password: %v", err)
	}
}

func TestNewPassword_nonASCII(t *testing.T) {
	cases := []string{
		"Pass1!あいう",
		"Pass1!\x00hidden",
		"Pass1!\x7fdelete",
		"パスワード1!ABC",
	}
	for _, tc := range cases {
		_, err := user.NewPassword(tc)
		if !errors.Is(err, user.ErrPasswordNotASCII) {
			t.Errorf("NewPassword(%q) error = %v, want ErrPasswordNotASCII", tc, err)
		}
	}
}

func TestNewPasswordFromHash_returnsHash(t *testing.T) {
	p := user.NewPasswordFromHash("$2a$10$somehashvalue")
	if p.Hash() != "$2a$10$somehashvalue" {
		t.Errorf("Hash() = %q, want %q", p.Hash(), "$2a$10$somehashvalue")
	}
}
