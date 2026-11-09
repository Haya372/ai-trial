package tag

//go:generate go tool mockgen -destination=generated/repository.go -package=mock github.com/Haya372/ai-trial/backend/domain/tag Repository

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	// Create returns ErrTagNameTaken if the owner already has a tag with the
	// same name, compared case-insensitively.
	Create(ctx context.Context, t Tag) (Tag, error)
	// FindByID returns the tag with the given ID, or ErrTagNotFound if it does not exist.
	FindByID(ctx context.Context, id uuid.UUID) (Tag, error)
	// Update returns ErrTagNotFound if the tag no longer exists, and
	// ErrTagNameTaken if another tag of the owner has the same name, compared
	// case-insensitively.
	Update(ctx context.Context, t Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	// CountByUserID returns how many tags userID owns, for the MaxTagsPerUser check.
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	// CountEventsByTagID returns how many events the tag is attached to, so a
	// command can answer with the same eventCount that GET /tags reports.
	CountEventsByTagID(ctx context.Context, tagID uuid.UUID) (int, error)
	// LockOwner locks the owner's user row (SELECT ... FOR UPDATE) so that
	// concurrent creates by the same user serialize their MaxTagsPerUser check.
	// It must be called inside a transaction (ADR-024).
	LockOwner(ctx context.Context, userID uuid.UUID) error
	// FindByIDsForShare returns the tags with the given IDs and locks their
	// rows (SELECT ... FOR SHARE) so they cannot be deleted before an event
	// references them. IDs with no tag are omitted from the result rather than
	// reported as an error. It must be called inside a transaction (ADR-024).
	FindByIDsForShare(ctx context.Context, ids []uuid.UUID) ([]Tag, error)
}
