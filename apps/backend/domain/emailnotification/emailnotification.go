package emailnotification

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxAttempts is the first attempt plus two retries (SPEC-010).
	MaxAttempts = 3
	// ClaimLease is how long a claim keeps other workers away. A claim
	// whose result is not recorded by then is treated as a stopped worker.
	ClaimLease = 60 * time.Second
)

// retryDelays[i] is the delay after the (i+1)th attempt fails transiently.
var retryDelays = [MaxAttempts - 1]time.Duration{10 * time.Second, 20 * time.Second}

// These are not DomainErrors: the worker only transitions notifications it
// has just claimed, so hitting them means a bug, not bad input.
var (
	ErrNotPending = errors.New("email notification is not pending")
	ErrNotClaimed = errors.New("email notification has not been claimed")

	errBrokenInvariant = errors.New("email notification invariant broken")
)

// EmailNotification is the aggregate root for the delivery record of one
// email. Transition methods return a new value and leave the receiver as
// is. See docs/domain/emailnotification.md for the state machine.
type EmailNotification interface {
	ID() uuid.UUID
	UserID() uuid.UUID
	Recipient() RecipientEmail
	Kind() EmailKind
	Status() EmailStatus
	// AttemptCount counts claims, so it includes an attempt whose worker
	// stopped before recording a result.
	AttemptCount() int
	RequestedAt() time.Time
	// NextAttemptAt is set only while pending.
	NextAttemptAt() (time.Time, bool)
	// SentAt is set only when sent.
	SentAt() (time.Time, bool)
	// LastFailureCode is the failure of the latest attempt, if it failed.
	LastFailureCode() (FailureCode, bool)
	// TemplateParams is empty once the notification is sent or failed
	// (ADR-039).
	TemplateParams() TemplateParams

	// Claim starts an attempt: it increments the attempt count and pushes
	// NextAttemptAt to the end of the lease. If MaxAttempts claims have
	// already been made, the last attempt's worker stopped without a
	// result, so it fails with worker_interrupted instead.
	Claim(now time.Time) (EmailNotification, error)
	MarkSent(now time.Time) (EmailNotification, error)
	// RecordFailure schedules a retry for a transient failure while
	// attempts remain, and fails the notification otherwise.
	RecordFailure(code FailureCode, now time.Time) (EmailNotification, error)
}

type notificationEntity struct {
	id              uuid.UUID
	userID          uuid.UUID
	recipient       RecipientEmail
	kind            EmailKind
	status          EmailStatus
	attemptCount    int
	requestedAt     time.Time
	nextAttemptAt   *time.Time
	sentAt          *time.Time
	lastFailureCode *FailureCode
	params          TemplateParams
}

// New creates a pending notification that is due immediately. The caller
// checks the kind and params against TemplateRegistry first.
func New(
	id, userID uuid.UUID, recipient RecipientEmail, kind EmailKind, params TemplateParams, now time.Time,
) EmailNotification {
	return &notificationEntity{
		id:            id,
		userID:        userID,
		recipient:     recipient,
		kind:          kind,
		status:        StatusPending,
		requestedAt:   now,
		nextAttemptAt: &now,
		params:        params,
	}
}

type RestoreParams struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Recipient       RecipientEmail
	Kind            EmailKind
	Status          EmailStatus
	AttemptCount    int
	RequestedAt     time.Time
	NextAttemptAt   *time.Time
	SentAt          *time.Time
	LastFailureCode *FailureCode
	TemplateParams  TemplateParams
}

// Restore rebuilds a stored notification and rejects one that breaks the
// aggregate's invariants.
func Restore(p RestoreParams) (EmailNotification, error) {
	n := &notificationEntity{
		id:              p.ID,
		userID:          p.UserID,
		recipient:       p.Recipient,
		kind:            p.Kind,
		status:          p.Status,
		attemptCount:    p.AttemptCount,
		requestedAt:     p.RequestedAt,
		nextAttemptAt:   p.NextAttemptAt,
		sentAt:          p.SentAt,
		lastFailureCode: p.LastFailureCode,
		params:          p.TemplateParams,
	}
	if err := n.checkInvariants(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *notificationEntity) checkInvariants() error {
	if _, err := NewEmailStatus(string(n.status)); err != nil {
		return err
	}
	var broken string
	switch {
	case n.attemptCount < 0:
		broken = "attempt count is negative"
	case (n.status == StatusPending) != (n.nextAttemptAt != nil):
		broken = "next attempt time must be set exactly while pending"
	case n.status == StatusSent && n.sentAt == nil:
		broken = "sent notification has no sent time"
	case n.status == StatusFailed && n.lastFailureCode == nil:
		broken = "failed notification has no failure code"
	case n.status.IsTerminal() && !n.params.IsEmpty():
		broken = "terminal notification still has template params"
	default:
		return nil
	}
	return fmt.Errorf("%w: %s (id %s)", errBrokenInvariant, broken, n.id)
}

func (n *notificationEntity) ID() uuid.UUID             { return n.id }
func (n *notificationEntity) UserID() uuid.UUID         { return n.userID }
func (n *notificationEntity) Recipient() RecipientEmail { return n.recipient }
func (n *notificationEntity) Kind() EmailKind           { return n.kind }
func (n *notificationEntity) Status() EmailStatus       { return n.status }
func (n *notificationEntity) AttemptCount() int         { return n.attemptCount }
func (n *notificationEntity) RequestedAt() time.Time    { return n.requestedAt }
func (n *notificationEntity) TemplateParams() TemplateParams {
	return n.params
}

func (n *notificationEntity) NextAttemptAt() (time.Time, bool) { return deref(n.nextAttemptAt) }
func (n *notificationEntity) SentAt() (time.Time, bool)        { return deref(n.sentAt) }
func (n *notificationEntity) LastFailureCode() (FailureCode, bool) {
	return deref(n.lastFailureCode)
}

func deref[T any](p *T) (T, bool) {
	if p == nil {
		var zero T
		return zero, false
	}
	return *p, true
}

func (n *notificationEntity) Claim(now time.Time) (EmailNotification, error) {
	if n.status != StatusPending {
		return nil, fmt.Errorf("%w: %s", ErrNotPending, n.status)
	}
	if n.attemptCount >= MaxAttempts {
		return n.fail(FailureWorkerInterrupted), nil
	}
	copied := *n
	copied.attemptCount++
	leaseEnd := now.Add(ClaimLease)
	copied.nextAttemptAt = &leaseEnd
	return &copied, nil
}

func (n *notificationEntity) MarkSent(now time.Time) (EmailNotification, error) {
	if err := n.checkClaimed(); err != nil {
		return nil, err
	}
	copied := *n
	copied.status = StatusSent
	copied.sentAt = &now
	copied.nextAttemptAt = nil
	copied.params = TemplateParams{}
	return &copied, nil
}

func (n *notificationEntity) RecordFailure(code FailureCode, now time.Time) (EmailNotification, error) {
	if err := n.checkClaimed(); err != nil {
		return nil, err
	}
	if !code.IsTransient() || n.attemptCount >= MaxAttempts {
		return n.fail(code), nil
	}
	copied := *n
	copied.lastFailureCode = &code
	next := now.Add(retryDelays[n.attemptCount-1])
	copied.nextAttemptAt = &next
	return &copied, nil
}

func (n *notificationEntity) checkClaimed() error {
	if n.status != StatusPending {
		return fmt.Errorf("%w: %s", ErrNotPending, n.status)
	}
	if n.attemptCount == 0 {
		return ErrNotClaimed
	}
	return nil
}

func (n *notificationEntity) fail(code FailureCode) EmailNotification {
	copied := *n
	copied.status = StatusFailed
	copied.lastFailureCode = &code
	copied.nextAttemptAt = nil
	copied.params = TemplateParams{}
	return &copied
}
