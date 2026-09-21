package user

import (
	"fmt"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/Haya372/ai-trial/backend/domain"
)

type Password struct {
	plain string
	hash  string
}

const (
	CodePasswordTooShort               = "TOO_SHORT"
	CodePasswordTooLong                = "TOO_LONG"
	CodePasswordNotASCII               = "INVALID_CHARACTER" //nolint:gosec
	CodePasswordInsufficientComplexity = "INSUFFICIENT_COMPLEXITY"
)

var (
	ErrPasswordTooShort = domain.NewDomainError(
		CodePasswordTooShort, "Password must be at least 8 characters")
	ErrPasswordTooLong = domain.NewDomainError(
		CodePasswordTooLong, "Password must be at most 72 characters")
	ErrPasswordNotASCII = domain.NewDomainError(
		CodePasswordNotASCII, "Password must contain only ASCII printable characters")
	ErrPasswordInsufficientComplexity = domain.NewDomainError(
		CodePasswordInsufficientComplexity, "Password must contain uppercase, lowercase, digit, and symbol")
)

func NewPassword(plain string) (Password, error) {
	for _, b := range []byte(plain) {
		if b < 0x20 || b > 0x7E {
			return Password{}, fmt.Errorf("%w", ErrPasswordNotASCII)
		}
	}
	if len(plain) < 8 {
		return Password{}, fmt.Errorf("%w", ErrPasswordTooShort)
	}
	// bcrypt silently truncates input beyond 72 bytes; we reject to avoid
	// two different passwords producing the same hash.
	if len(plain) > 72 {
		return Password{}, fmt.Errorf("%w", ErrPasswordTooLong)
	}
	if err := checkComplexity(plain); err != nil {
		return Password{}, err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return Password{}, fmt.Errorf("failed to hash password: %w", err)
	}
	return Password{plain: plain, hash: string(hashed)}, nil
}

func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Hash() string {
	return p.hash
}

// checkComplexity checks character class coverage. Called only after NewPassword
// has validated that p consists solely of ASCII printable characters (0x20-0x7E).
func checkComplexity(p string) error {
	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range p {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSymbol {
		return fmt.Errorf("%w", ErrPasswordInsufficientComplexity)
	}
	return nil
}
