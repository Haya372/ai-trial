package event

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

// RecurrenceInput carries the raw recurrence fields from the API request.
// All string values are unvalidated; buildRule validates them.
type RecurrenceInput struct {
	PatternType string
	MonthlyNth  *string
	EndType     string
	UntilDate   *string // YYYY-MM-DD
	Count       *int
	Timezone    string
}

type CreateEventInput struct {
	Title       string
	Description string
	StartAt     time.Time
	EndAt       time.Time
	Location    string
	URL         string
	// TagIDs lists tag IDs to attach in order. nil and empty both mean no tags.
	TagIDs []uuid.UUID
	// Recurrence is nil for a one-off event.
	Recurrence *RecurrenceInput
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
	// Series is nil for a one-off event.
	Series rs.RecurringEventSeries
}

type CreateEventCommand struct {
	eventRepo  domainevent.Repository
	seriesRepo rs.Repository
	tagRepo    tag.Repository
	txManager  usecase.TransactionManager
}

func NewCreateEventCommand(
	r domainevent.Repository,
	sr rs.Repository,
	tr tag.Repository,
	tx usecase.TransactionManager,
) *CreateEventCommand {
	return &CreateEventCommand{eventRepo: r, seriesRepo: sr, tagRepo: tr, txManager: tx}
}

func (c *CreateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in CreateEventInput,
) (CreateEventOutput, error) {
	// Validate the first occurrence (title, endAt > startAt) before any path.
	firstOcc, err := domainevent.New(
		uuid.New(), userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
	)
	if err != nil {
		return CreateEventOutput{}, err
	}

	// Validate tag count and duplicates before entering the transaction.
	tags, err := domainevent.NewEventTags(in.TagIDs)
	if err != nil {
		return CreateEventOutput{}, toTagValidationError(err)
	}

	if in.Recurrence == nil {
		return c.createOneOff(ctx, userID, in, firstOcc, tags)
	}
	return c.createRecurring(ctx, userID, in, tags)
}

// createOneOff saves a single non-recurring event.
func (c *CreateEventCommand) createOneOff(
	ctx context.Context,
	userID uuid.UUID,
	in CreateEventInput,
	firstOcc domainevent.Event,
	eventTags []domainevent.EventTag,
) (CreateEventOutput, error) {
	var saved domainevent.Event
	var tagViews []EventTagView
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		var err error
		tagViews, err = c.lockOwnedTags(ctx, in.TagIDs, userID)
		if err != nil {
			return err
		}
		tagged, err := firstOcc.WithTags(eventTags)
		if err != nil {
			return err
		}
		saved, err = c.eventRepo.Create(ctx, tagged)
		return err
	}); err != nil {
		return CreateEventOutput{}, fmt.Errorf("create event: %w", err)
	}
	return CreateEventOutput{Event: saved, Tags: tagViews}, nil
}

// createRecurring validates the recurrence rule, expands occurrences and saves
// the series together with all occurrences in a single transaction.
func (c *CreateEventCommand) createRecurring(
	ctx context.Context,
	userID uuid.UUID,
	in CreateEventInput,
	eventTags []domainevent.EventTag,
) (CreateEventOutput, error) {
	rule, err := buildRule(*in.Recurrence)
	if err != nil {
		return CreateEventOutput{}, err
	}

	occs, err := rs.ExpandOccurrences(in.StartAt, in.EndAt, rule, nil)
	if err != nil {
		return CreateEventOutput{}, toExpandValidationError(err)
	}

	series, err := rs.New(uuid.New(), userID, rule, nil)
	if err != nil {
		return CreateEventOutput{}, fmt.Errorf("create series: %w", err)
	}

	var savedSeries rs.RecurringEventSeries
	var tagViews []EventTagView
	var first domainevent.Event
	if err := c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		tagViews, err = c.lockOwnedTags(ctx, in.TagIDs, userID)
		if err != nil {
			return err
		}
		savedSeries, err = c.seriesRepo.Create(ctx, series)
		if err != nil {
			return err
		}
		events, err := c.buildOccurrences(userID, in, savedSeries.ID(), occs, eventTags)
		if err != nil {
			return err
		}
		if err := c.seriesRepo.CreateOccurrences(ctx, events); err != nil {
			return err
		}
		// The bulk insert returns no rows, so read the first occurrence back
		// to return the stored values (e.g. microsecond timestamps) like the
		// one-off path does.
		first, err = c.eventRepo.FindByID(ctx, events[0].ID())
		return err
	}); err != nil {
		return CreateEventOutput{}, fmt.Errorf("create recurring event: %w", err)
	}

	return CreateEventOutput{Event: first, Tags: tagViews, Series: savedSeries}, nil
}

// buildOccurrences constructs one Event per occurrence, attaching the series ID
// and the tags to each.
func (c *CreateEventCommand) buildOccurrences(
	userID uuid.UUID,
	in CreateEventInput,
	seriesID uuid.UUID,
	occs []rs.Occurrence,
	eventTags []domainevent.EventTag,
) ([]domainevent.Event, error) {
	events := make([]domainevent.Event, len(occs))
	for i, occ := range occs {
		e, err := domainevent.New(
			uuid.New(), userID, in.Title, in.Description,
			occ.StartAt, occ.EndAt, in.Location, in.URL,
		)
		if err != nil {
			return nil, fmt.Errorf("create occurrence %d: %w", i, err)
		}
		e, err = e.WithSeriesID(seriesID).WithTags(eventTags)
		if err != nil {
			return nil, fmt.Errorf("attach tags to occurrence %d: %w", i, err)
		}
		events[i] = e
	}
	return events, nil
}

// lockOwnedTags locks the requested tags (FOR SHARE) so they cannot be
// deleted before the event rows refer to them, and checks that the caller
// owns them. Call it inside the transaction that saves the event.
func (c *CreateEventCommand) lockOwnedTags(
	ctx context.Context, tagIDs []uuid.UUID, userID uuid.UUID,
) ([]EventTagView, error) {
	if len(tagIDs) == 0 {
		return nil, nil
	}
	locked, err := c.tagRepo.FindByIDsForShare(ctx, tagIDs)
	if err != nil {
		return nil, err
	}
	return verifyAndBuildTagViews(tagIDs, locked, userID)
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
