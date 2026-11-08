package tag

import (
	"fmt"
	"strings"

	"github.com/Haya372/ai-trial/backend/domain"
)

type TagName string

const (
	CodeTagNameEmpty   = "REQUIRED"
	CodeTagNameTooLong = "TOO_LONG"
	maxTagNameLength   = 20
)

var (
	ErrTagNameEmpty   = domain.NewDomainError(CodeTagNameEmpty, "Tag name is required")
	ErrTagNameTooLong = domain.NewDomainError(CodeTagNameTooLong, "Tag name must be at most 20 characters")
)

// NewTagName trims surrounding whitespace before validating, so the stored
// name and the length check both use the trimmed value (SPEC-008).
func NewTagName(s string) (TagName, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", fmt.Errorf("%w", ErrTagNameEmpty)
	}
	if len([]rune(trimmed)) > maxTagNameLength {
		return "", fmt.Errorf("%w", ErrTagNameTooLong)
	}
	return TagName(trimmed), nil
}
