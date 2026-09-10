package user

import (
	"fmt"
	"regexp"
)

type Email string

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NewEmail(s string) (Email, error) {
	if !emailRegex.MatchString(s) {
		return "", fmt.Errorf("invalid email address: %q", s)
	}
	return Email(s), nil
}
