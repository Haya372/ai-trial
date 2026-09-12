package user_test

import (
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

func TestNewEmail_valid(t *testing.T) {
	cases := []string{
		"user@example.com",
		"user+tag@sub.example.co.jp",
		"a@b.io",
	}
	for _, addr := range cases {
		email, err := user.NewEmail(addr)
		if err != nil {
			t.Errorf("NewEmail(%q) unexpected error: %v", addr, err)
		}
		if string(email) != addr {
			t.Errorf("NewEmail(%q) = %q, want %q", addr, email, addr)
		}
	}
}

func TestNewEmail_invalid(t *testing.T) {
	cases := []string{
		"",
		"notanemail",
		"missing@",
		"@nodomain.com",
		"spaces in@email.com",
	}
	for _, addr := range cases {
		_, err := user.NewEmail(addr)
		if err == nil {
			t.Errorf("NewEmail(%q) expected error, got nil", addr)
		}
	}
}
