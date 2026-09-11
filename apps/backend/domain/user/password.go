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
	ErrPasswordTooLong  = errors.New("password too long: maximum 128 characters")
)

func NewPassword(plain string) (Password, error) {
	if len(plain) < 8 {
		return Password{}, fmt.Errorf("%w", ErrPasswordTooShort)
	}
	if len(plain) > 128 {
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
