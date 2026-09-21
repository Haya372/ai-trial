package user

import (
	"fmt"

	"github.com/Haya372/ai-trial/backend/domain"
)

const (
	CodeDisplayNameTooLong = "TOO_LONG"
	maxDisplayNameLength   = 50
)

var ErrDisplayNameTooLong error = &domain.DomainError{
	Code:    CodeDisplayNameTooLong,
	Message: "Display name must be at most 50 characters",
}

func NewDisplayName(s string) (string, error) {
	if len([]rune(s)) > maxDisplayNameLength {
		return "", fmt.Errorf("%w", ErrDisplayNameTooLong)
	}
	return s, nil
}
