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
	// 72-char password that satisfies all complexity requirements
	p72 := "SecurePass1!" + strings.Repeat("a", 60)
	_, err := user.NewPassword(p72)
	if err != nil {
		t.Errorf("NewPassword() unexpected error for 72-char password: %v", err)
	}
}

func TestNewPassword_insufficientComplexity(t *testing.T) {
	cases := []string{
		"alllowercase1!", // no uppercase
		"ALLUPPERCASE1!", // no lowercase
		"NoDigitsHere!!", // no digit
		"NoSymbols1234A", // no symbol
	}
	for _, tc := range cases {
		_, err := user.NewPassword(tc)
		if !errors.Is(err, user.ErrPasswordInsufficientComplexity) {
			t.Errorf("NewPassword(%q) error = %v, want ErrPasswordInsufficientComplexity", tc, err)
		}
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

func TestCompareDummyPassword_returnsMismatchForArbitraryPassword(t *testing.T) {
	p, err := user.NewPassword("SecurePass1!")
	if err != nil {
		t.Fatalf("NewPassword() unexpected error: %v", err)
	}
	if err := user.CompareDummyPassword(p); !errors.Is(err, user.ErrPasswordMismatch) {
		t.Errorf("CompareDummyPassword() error = %v, want ErrPasswordMismatch", err)
	}
}

// NewLoginPassword tests

func TestNewLoginPassword_valid_hashPanics(t *testing.T) {
	p, err := user.NewLoginPassword("SecurePass1!")
	if err != nil {
		t.Fatalf("NewLoginPassword() unexpected error: %v", err)
	}
	// A login-only Password must never be persisted. Hash() panics instead of
	// returning an empty string, so misuse (e.g. passing it to Repository.Create)
	// fails loudly instead of silently locking the account out.
	defer func() {
		if recover() == nil {
			t.Error("Hash() did not panic for a password built via NewLoginPassword")
		}
	}()
	p.Hash()
}

func TestNewLoginPassword_tooShort(t *testing.T) {
	_, err := user.NewLoginPassword("short")
	if !errors.Is(err, user.ErrPasswordTooShort) {
		t.Errorf("NewLoginPassword() error = %v, want ErrPasswordTooShort", err)
	}
}

func TestNewLoginPassword_tooLong(t *testing.T) {
	_, err := user.NewLoginPassword(strings.Repeat("a", 73))
	if !errors.Is(err, user.ErrPasswordTooLong) {
		t.Errorf("NewLoginPassword() error = %v, want ErrPasswordTooLong", err)
	}
}

func TestNewLoginPassword_maxLength(t *testing.T) {
	// 72-char password that satisfies all complexity requirements
	p72 := "SecurePass1!" + strings.Repeat("a", 60)
	_, err := user.NewLoginPassword(p72)
	if err != nil {
		t.Errorf("NewLoginPassword() unexpected error for 72-char password: %v", err)
	}
}

func TestNewLoginPassword_nonASCII(t *testing.T) {
	cases := []string{
		"Pass1!あいう",
		"Pass1!\x00hidden",
		"Pass1!\x7fdelete",
		"パスワード1!ABC",
	}
	for _, tc := range cases {
		_, err := user.NewLoginPassword(tc)
		if !errors.Is(err, user.ErrPasswordNotASCII) {
			t.Errorf("NewLoginPassword(%q) error = %v, want ErrPasswordNotASCII", tc, err)
		}
	}
}

func TestNewLoginPassword_insufficientComplexity(t *testing.T) {
	cases := []string{
		"alllowercase1!", // no uppercase
		"ALLUPPERCASE1!", // no lowercase
		"NoDigitsHere!!", // no digit
		"NoSymbols1234A", // no symbol
	}
	for _, tc := range cases {
		_, err := user.NewLoginPassword(tc)
		if !errors.Is(err, user.ErrPasswordInsufficientComplexity) {
			t.Errorf("NewLoginPassword(%q) error = %v, want ErrPasswordInsufficientComplexity", tc, err)
		}
	}
}

func TestCompareDummyPassword_withLoginPassword(t *testing.T) {
	p, err := user.NewLoginPassword("SecurePass1!")
	if err != nil {
		t.Fatalf("NewLoginPassword() unexpected error: %v", err)
	}
	if err := user.CompareDummyPassword(p); !errors.Is(err, user.ErrPasswordMismatch) {
		t.Errorf("CompareDummyPassword() error = %v, want ErrPasswordMismatch", err)
	}
}
