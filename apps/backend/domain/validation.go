package domain

import "fmt"

type ValidationDetail struct {
	Field   string
	Code    string
	Message string
}

type ValidationError struct {
	Details []ValidationDetail
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %d field(s) invalid", len(e.Details))
}
