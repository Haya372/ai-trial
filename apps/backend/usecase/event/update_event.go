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
	// TagIDs, when nil, keeps existing tags unchanged. When non-nil the event's
	// tags are replaced with this slice (empty slice removes all tags).
	TagIDs *[]uuid.UUID
}

type UpdateEventOutput struct {
	Event domainevent.Event
	Tags  []EventTagView
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
) (UpdateEventOutput, error) {
	// Body validation happens before entering the transaction (per design).
	if _, err := domainevent.New(
		in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
	); err != nil {
		return UpdateEventOutput{}, err
	}

	// Validate replacement tag count/duplicates before the transaction as well.
	if in.TagIDs != nil {
		if _, err := domainevent.NewEventTags(*in.TagIDs); err != nil {
			return UpdateEventOutput{}, toTagValidationError(err)
		}
	}

	var updated domainevent.Event
	var tagViews []EventTagView
	err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		existing, err := domainevent.FindOwnedForUpdate(ctx, c.repo, c.logger, in.ID, userID)
		if err != nil {
			return err
		}

		newTags, views, err := c.determineTagsInTx(ctx, userID, existing, in.TagIDs)
		if err != nil {
			return err
		}
		tagViews = views

		updated, err = domainevent.New(
			in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
		)
		if err != nil {
			return err
		}
		updated, err = updated.WithTags(newTags)
		if err != nil {
			return err
		}
		if err := c.repo.Update(ctx, updated); err != nil {
			return fmt.Errorf("update event: %w", err)
		}
		return nil
	})
	if err != nil {
		return UpdateEventOutput{}, err
	}
	return UpdateEventOutput{Event: updated, Tags: tagViews}, nil
}

// determineTagsInTx selects the new tag set for the update:
//   - non-nil TagIDs: replace with the requested tags (may be empty to remove all)
//   - nil TagIDs: re-lock existing tags and drop any that were deleted concurrently
func (c *UpdateEventCommand) determineTagsInTx(
	ctx context.Context,
	userID uuid.UUID,
	existing domainevent.Event,
	tagIDs *[]uuid.UUID,
) ([]domainevent.EventTag, []EventTagView, error) {
	if tagIDs != nil {
		return c.replaceTagsInTx(ctx, userID, *tagIDs)
	}
	return c.keepTagsInTx(ctx, existing.Tags())
}

// replaceTagsInTx builds the EventTag slice for an explicit tag replacement.
func (c *UpdateEventCommand) replaceTagsInTx(
	ctx context.Context, userID uuid.UUID, tagIDs []uuid.UUID,
) ([]domainevent.EventTag, []EventTagView, error) {
	newEventTags, _ := domainevent.NewEventTags(tagIDs) // already validated outside tx
	if len(tagIDs) == 0 {
		return newEventTags, nil, nil
	}
	locked, err := c.tagRepo.FindByIDsForShare(ctx, tagIDs)
	if err != nil {
		return nil, nil, err
	}
	tagViews, err := verifyAndBuildTagViews(tagIDs, locked, userID)
	if err != nil {
		return nil, nil, err
	}
	return newEventTags, tagViews, nil
}

// keepTagsInTx re-locks the existing tag rows (FOR SHARE) and drops any that
// have been deleted since the event was last read.
func (c *UpdateEventCommand) keepTagsInTx(
	ctx context.Context, existing []domainevent.EventTag,
) ([]domainevent.EventTag, []EventTagView, error) {
	keptTags, err := c.lockKeptTags(ctx, existing)
	if err != nil {
		return nil, nil, err
	}
	return keptTags, eventTagsToViews(keptTags), nil
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

// eventTagsToViews converts EventTag slice to EventTagView slice. The
// nil-TagIDs path uses existing tags; the handler builds the full response
// from out.Event.Tags() for the nil-TagIDs (kept-tags) path.
func eventTagsToViews(tags []domainevent.EventTag) []EventTagView {
	views := make([]EventTagView, len(tags))
	for i, t := range tags {
		views[i] = EventTagView{ID: t.TagID()}
	}
	return views
}
