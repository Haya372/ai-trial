package emailnotification

import (
	"errors"
	"fmt"
)

// FailureCode records why an attempt failed. It replaces the error text the
// destination returns, which may contain the recipient address (SPEC-010).
type FailureCode string

const (
	FailureConnectionFailed     FailureCode = "connection_failed"
	FailureTimeout              FailureCode = "timeout"
	FailureRateLimited          FailureCode = "rate_limited"
	FailureTemporaryRejection   FailureCode = "temporary_rejection"
	FailureInvalidRecipient     FailureCode = "invalid_recipient"
	FailureRecipientRejected    FailureCode = "recipient_rejected"
	FailureAuthFailed           FailureCode = "auth_failed"
	FailurePermanentRejection   FailureCode = "permanent_rejection"
	FailureTemplateRenderFailed FailureCode = "template_render_failed"
	FailureWorkerInterrupted    FailureCode = "worker_interrupted"
)

var errUnknownFailureCode = errors.New("unknown email failure code")

func NewFailureCode(s string) (FailureCode, error) {
	switch c := FailureCode(s); c {
	case FailureConnectionFailed, FailureTimeout, FailureRateLimited, FailureTemporaryRejection,
		FailureInvalidRecipient, FailureRecipientRejected, FailureAuthFailed, FailurePermanentRejection,
		FailureTemplateRenderFailed, FailureWorkerInterrupted:
		return c, nil
	}
	return "", fmt.Errorf("%w: %q", errUnknownFailureCode, s)
}

// IsTransient reports whether the failure is retried. Every code not listed
// here is permanent and fails the notification at once.
func (c FailureCode) IsTransient() bool {
	switch c {
	case FailureConnectionFailed, FailureTimeout, FailureRateLimited, FailureTemporaryRejection:
		return true
	}
	return false
}
