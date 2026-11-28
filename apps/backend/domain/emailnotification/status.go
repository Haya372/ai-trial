package emailnotification

import (
	"errors"
	"fmt"
)

type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
)

// errUnknownStatus is not a DomainError: statuses come from storage, never
// from user input, so an unknown one means corrupted data or a bug.
var errUnknownStatus = errors.New("unknown email status")

func NewEmailStatus(s string) (EmailStatus, error) {
	switch st := EmailStatus(s); st {
	case StatusPending, StatusSent, StatusFailed:
		return st, nil
	}
	return "", fmt.Errorf("%w: %q", errUnknownStatus, s)
}

func (s EmailStatus) IsTerminal() bool {
	return s == StatusSent || s == StatusFailed
}
