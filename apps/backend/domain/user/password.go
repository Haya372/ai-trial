package user

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	hash string
}

var (
	ErrPasswordTooShort = errors.New("password too short: minimum 8 characters")
	ErrPasswordTooLong  = errors.New("password too long: maximum 72 characters")
	ErrPasswordNotASCII = errors.New("password must contain only ASCII printable characters (0x20-0x7E)")
)

func NewPassword(plain string) (Password, error) {
	for i := 0; i < len(plain); i++ {
		if plain[i] < 0x20 || plain[i] > 0x7E {
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
	return Password{hash: string(hashed)}, nil
}

func NewPasswordFromHash(hash string) Password {
	return Password{hash: hash}
}

func (p Password) Hash() string {
	return p.hash
}
