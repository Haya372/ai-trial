package user

import (
	"fmt"
	"regexp"

	"github.com/Haya372/ai-trial/backend/domain"
)

type Email string

const CodeInvalidEmailFormat = "INVALID_FORMAT"

var (
	ErrInvalidEmail = &domain.DomainError{Code: CodeInvalidEmailFormat, Message: "Invalid email format"}
	emailRegex      = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

func NewEmail(s string) (Email, error) {
	if !emailRegex.MatchString(s) {
		return "", fmt.Errorf("%w: %q", ErrInvalidEmail, s)
	}
	return Email(s), nil
}
