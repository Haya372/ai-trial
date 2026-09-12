package user

import (
	"errors"
	"fmt"
	"regexp"
)

type Email string

var (
	ErrInvalidEmail = errors.New("invalid email address")
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func NewEmail(s string) (Email, error) {
	if !emailRegex.MatchString(s) {
		return "", fmt.Errorf("%w: %q", ErrInvalidEmail, s)
	}
	return Email(s), nil
}
