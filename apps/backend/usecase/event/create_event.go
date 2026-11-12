package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

type CreateEventInput struct {
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
	// TagIDs lists tag IDs to attach in order. nil and empty both mean no tags.
	TagIDs []uuid.UUID
}

// EventTagView carries tag display data for Command output (Create/Update).
type EventTagView struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type CreateEventOutput struct {
	Event domainevent.Event
	Tags  []EventTagView
}

type CreateEventCommand struct {
	eventRepo domainevent.Repository
	tagRepo   tag.Repository
	txManager usecase.TransactionManager
}

func NewCreateEventCommand(
	r domainevent.Repository, tr tag.Repository, tx usecase.TransactionManager,
) *CreateEventCommand {
	return &CreateEventCommand{eventRepo: r, tagRepo: tr, txManager: tx}
}

func (c *CreateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateEventInput,
) (CreateEventOutput, error) {
	id := uuid.New()
	e, err := domainevent.New(id, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL)
	if err != nil {
		return CreateEventOutput{}, err
	}

	// Validate tag count and duplicates before entering the transaction.
	eventTags, err := domainevent.NewEventTags(in.TagIDs)
	if err != nil {
		return CreateEventOutput{}, toTagValidationError(err)
	}

	var saved domainevent.Event
	var tagViews []EventTagView
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		if len(in.TagIDs) > 0 {
			locked, err := c.tagRepo.FindByIDsForShare(ctx, in.TagIDs)
			if err != nil {
				return err
			}
			tagViews, err = verifyAndBuildTagViews(in.TagIDs, locked, userID)
			if err != nil {
				return err
			}
			e, err = e.WithTags(eventTags)
			if err != nil {
				return err
			}
		}
		saved, err = c.eventRepo.Create(ctx, e)
		return err
	}); err != nil {
		return CreateEventOutput{}, fmt.Errorf("create event: %w", err)
	}
	return CreateEventOutput{Event: saved, Tags: tagViews}, nil
}

// verifyAndBuildTagViews checks that all requested tag IDs exist and are owned
// by userID, returning a ValidationError{tagIds: NOT_FOUND} otherwise.
func verifyAndBuildTagViews(tagIDs []uuid.UUID, locked []tag.Tag, userID uuid.UUID) ([]EventTagView, error) {
	byID := make(map[uuid.UUID]tag.Tag, len(locked))
	for _, t := range locked {
		byID[t.ID()] = t
	}
	views := make([]EventTagView, 0, len(tagIDs))
	for _, id := range tagIDs {
		t, ok := byID[id]
		if !ok || t.UserID() != userID {
			return nil, &domain.ValidationError{Details: []domain.ValidationDetail{{
				Field:   "tagIds",
				Code:    domainevent.CodeEventTagForeign,
				Message: "Tag not found",
			}}}
		}
		views = append(views, EventTagView{
			ID:    t.ID(),
			Name:  string(t.Name()),
			Color: string(t.Color()),
		})
	}
	return views, nil
}

// toTagValidationError converts an EventTag domain error into a ValidationError
// scoped to the "tagIds" field. Only ErrEventTagLimitExceeded and
// ErrEventTagDuplicated are expected here; other errors are passed through as-is.
func toTagValidationError(err error) error {
	if de, ok := errors.AsType[*domain.DomainError](err); ok {
		switch de.Code() {
		case domainevent.CodeEventTagLimitExceeded, domainevent.CodeEventTagDuplicated:
			return &domain.ValidationError{Details: []domain.ValidationDetail{{
				Field:   "tagIds",
				Code:    de.Code(),
				Message: de.Message(),
			}}}
		}
	}
	return err
}
