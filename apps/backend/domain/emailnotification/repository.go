package emailnotification

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/emailnotification Repository

import (
	"context"
	"time"
)

// Repository stores notifications for the outbox (ADR-038).
type Repository interface {
	// Create saves a new pending notification. Call it inside the
	// requesting use case's transaction so the business data and the
	// notification commit together.
	Create(ctx context.Context, n EmailNotification) error
	// FindDueForUpdate locks and returns one pending notification whose
	// NextAttemptAt is at or before now, skipping rows other workers have
	// locked (FOR UPDATE SKIP LOCKED). It returns nil, nil when none is
	// due. Call it inside a transaction and save the claim with SaveClaim
	// in the same transaction.
	FindDueForUpdate(ctx context.Context, now time.Time) (EmailNotification, error)
	// SaveClaim saves the result of EmailNotification.Claim for the row
	// FindDueForUpdate locked.
	SaveClaim(ctx context.Context, n EmailNotification) error
	// RecordResult saves the result of MarkSent or RecordFailure only if
	// the stored row is still pending with the same attempt count, that
	// is, no other worker has claimed it since. It reports whether a row
	// was updated; false means the result was discarded.
	RecordResult(ctx context.Context, n EmailNotification) (bool, error)
}
