package event

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

type UpdateEventInput struct {
	ID          uuid.UUID
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
}

type UpdateEventCommand struct {
	repo      domainevent.Repository
	tagRepo   tag.Repository
	txManager usecase.TransactionManager
	logger    *slog.Logger
}

func NewUpdateEventCommand(
	r domainevent.Repository, tr tag.Repository, tx usecase.TransactionManager, logger *slog.Logger,
) *UpdateEventCommand {
	return &UpdateEventCommand{repo: r, tagRepo: tr, txManager: tx, logger: logger}
}

func (c *UpdateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput,
) (domainevent.Event, error) {
	var updated domainevent.Event
	err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		// Lock the event so a concurrent edit cannot change its tags between
		// this read and the write below.
		existing, err := domainevent.FindOwnedForUpdate(ctx, c.repo, c.logger, in.ID, userID)
		if err != nil {
			return err
		}
		kept, err := c.lockKeptTags(ctx, existing.Tags())
		if err != nil {
			return err
		}

		updated, err = domainevent.New(
			in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
		)
		if err != nil {
			return err
		}
		// Update replaces the stored tags with updated.Tags(), so carry them
		// over or editing the event would drop its tags.
		updated, err = updated.WithTags(kept)
		if err != nil {
			return err
		}

		if err := c.repo.Update(ctx, updated); err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// lockKeptTags locks the rows of the tags being carried over (FOR SHARE) so
// they cannot be deleted before Update writes them back, and drops any tag
// already deleted since the event was read: writing it back would fail the
// event_tags foreign key.
func (c *UpdateEventCommand) lockKeptTags(
	ctx context.Context, tags []domainevent.EventTag,
) ([]domainevent.EventTag, error) {
	if len(tags) == 0 {
		return tags, nil
	}
	ids := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		ids[i] = t.TagID()
	}
	locked, err := c.tagRepo.FindByIDsForShare(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("lock event tags: %w", err)
	}

	present := make(map[uuid.UUID]struct{}, len(locked))
	for _, t := range locked {
		present[t.ID()] = struct{}{}
	}
	kept := make([]domainevent.EventTag, 0, len(tags))
	for _, t := range tags {
		if _, ok := present[t.TagID()]; ok {
			kept = append(kept, t)
		}
	}
	return kept, nil
}
