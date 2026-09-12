package user

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"github.com/Haya372/ai-trial/backend/domain"
)

type Password struct {
	plain string
	hash  string
}

const (
	CodePasswordTooShort = "TOO_SHORT"
	CodePasswordTooLong  = "TOO_LONG"
	CodePasswordNotASCII = "INVALID_CHARACTER" //nolint:gosec
)

var (
	ErrPasswordTooShort = &domain.DomainError{
		Code:    CodePasswordTooShort,
		Message: "Password must be at least 8 characters",
	}
	ErrPasswordTooLong = &domain.DomainError{
		Code:    CodePasswordTooLong,
		Message: "Password must be at most 72 characters",
	}
	ErrPasswordNotASCII = &domain.DomainError{
		Code:    CodePasswordNotASCII,
		Message: "Password must contain only ASCII printable characters",
	}
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
