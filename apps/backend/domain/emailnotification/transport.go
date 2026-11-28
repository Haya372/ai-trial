package emailnotification

//go:generate go tool mockgen -destination=generated/transport.go -package=mock github.com/Haya372/ai-trial/backend/domain/emailnotification EmailTransport

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Haya372/ai-trial/backend/domain/user"
)

// EmailTransport hands one email to a mail service. Implementations keep
// everything specific to the service and protocol, and translate every
// failure into a FailureCode instead of returning the service's error
// text, which may contain the recipient address (SPEC-010).
type EmailTransport interface {
	Send(ctx context.Context, email Email) SendResult
}

// Email is one plain-text email (ADR-041). Subject and Body contain the
// template params, so keep them out of logs and traces.
type Email struct {
	To      RecipientEmail
	From    Sender
	Subject string
	Body    string
}

// Sender is the single From name and address configured for the service.
type Sender struct {
	name    string
	address user.Email
}

var errInvalidSender = errors.New("invalid email sender")

func NewSender(name, address string) (Sender, error) {
	// The name goes into the From header, so a line break could inject
	// headers.
	if name == "" || strings.ContainsAny(name, "\r\n") {
		return Sender{}, fmt.Errorf("%w: name %q", errInvalidSender, name)
	}
	addr, err := user.NewEmail(address)
	if err != nil {
		return Sender{}, fmt.Errorf("%w: %w", errInvalidSender, err)
	}
	return Sender{name: name, address: addr}, nil
}

func (s Sender) Name() string    { return s.name }
func (s Sender) Address() string { return string(s.address) }

type SendOutcome int

const (
	OutcomeAccepted SendOutcome = iota + 1
	OutcomeTransientFailure
	OutcomePermanentFailure
)

// SendResult is either accepted or failed with a FailureCode. Whether a
// failure is transient follows from its code, so it is not stored apart.
//
// The zero value is a permanent failure with permanent_rejection, so a
// transport that returns SendResult{} by mistake never marks an email as
// sent, and retrying would not fix such a bug.
type SendResult struct {
	accepted bool
	failure  FailureCode
}

func SendAccepted() SendResult { return SendResult{accepted: true} }

func SendFailed(code FailureCode) SendResult { return SendResult{failure: code} }

func (r SendResult) Outcome() SendOutcome {
	switch {
	case r.accepted:
		return OutcomeAccepted
	case r.failure.IsTransient():
		return OutcomeTransientFailure
	default:
		return OutcomePermanentFailure
	}
}

func (r SendResult) FailureCode() (FailureCode, bool) {
	switch {
	case r.accepted:
		return "", false
	case r.failure == "":
		return FailurePermanentRejection, true
	default:
		return r.failure, true
	}
}
