package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/usecase"
)

// RecurrenceUpdate tells apart the three forms of recurrence in the request.
type RecurrenceUpdate struct {
	// Present is false when the request omitted recurrence: keep the current one.
	Present bool
	// Spec is nil for recurrence: null (stop repeating).
	Spec *RecurrenceInput
}

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
	TagIDs     *[]uuid.UUID
	Recurrence RecurrenceUpdate
	// Scope is "" when omitted. The command validates the value.
	Scope string
}

type UpdateEventOutput struct {
	Event domainevent.Event
	Tags  []EventTagView
	// Series is the series the returned event belongs to, nil for a one-off.
	Series rs.RecurringEventSeries
}

const (
	scopeSingle = "single"
	scopeFuture = "future"
	scopeAll    = "all"
)

// errScopeNotImplemented is not a DomainError, so it surfaces as a 500: no
// client sends these scopes before #407 implements them.
var errScopeNotImplemented = errors.New("scope future and all are not implemented yet")

// errSeriesMissing is not a DomainError: an event row refers to its series
// through a foreign key, so it comes from inconsistent data, not user input.
var errSeriesMissing = errors.New("event refers to a series that does not exist")

type UpdateEventCommand struct {
	repo       domainevent.Repository
	seriesRepo rs.Repository
	tagRepo    tag.Repository
	txManager  usecase.TransactionManager
	logger     *slog.Logger
}

func NewUpdateEventCommand(
	r domainevent.Repository,
	sr rs.Repository,
	tr tag.Repository,
	tx usecase.TransactionManager,
	logger *slog.Logger,
) *UpdateEventCommand {
	return &UpdateEventCommand{repo: r, seriesRepo: sr, tagRepo: tr, txManager: tx, logger: logger}
}

func (c *UpdateEventCommand) Execute(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput,
) (UpdateEventOutput, error) {
	rule, err := validateUpdateRequest(userID, in)
	if err != nil {
		return UpdateEventOutput{}, err
	}

	var out UpdateEventOutput
	err = c.txManager.RunInTx(ctx, func(ctx context.Context) error {
		existing, series, err := c.lockTarget(ctx, in.ID, userID)
		if err != nil {
			return err
		}
		if series == nil {
			out, err = c.updateOneOff(ctx, userID, in, existing, rule)
		} else {
			out, err = c.updateOccurrence(ctx, userID, in, existing, series, rule)
		}
		return err
	})
	if err != nil {
		return UpdateEventOutput{}, err
	}
	return out, nil
}

// validateUpdateRequest runs the checks that depend only on the request, so
// they happen before the transaction. It returns the parsed rule, or nil when
// the request sends no recurrence object.
func validateUpdateRequest(userID uuid.UUID, in UpdateEventInput) (*rs.RecurrenceRule, error) {
	if _, err := domainevent.New(
		in.ID, userID, in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL,
	); err != nil {
		return nil, err
	}
	if in.TagIDs != nil {
		if _, err := domainevent.NewEventTags(*in.TagIDs); err != nil {
			return nil, toTagValidationError(err)
		}
	}
	switch in.Scope {
	case "", scopeSingle, scopeFuture, scopeAll:
	default:
		return nil, validationDetail("scope", "INVALID_FORMAT", "scope must be single, future or all")
	}
	if in.Recurrence.Spec == nil {
		return nil, nil
	}
	rule, err := parseRule(*in.Recurrence.Spec)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// lockTarget locks the series row before the event row. "future" and "all"
// lock the series and then its occurrences, so taking the event first could
// deadlock against them. An unlocked read tells which series to lock.
func (c *UpdateEventCommand) lockTarget(
	ctx context.Context, id, userID uuid.UUID,
) (domainevent.Event, rs.RecurringEventSeries, error) {
	ev, err := domainevent.FindOwned(ctx, c.repo, c.logger, id, userID)
	if err != nil {
		return nil, nil, err
	}
	var series rs.RecurringEventSeries
	if seriesID, ok := ev.SeriesID(); ok {
		// A missing series is decided by the locked read below: a concurrent
		// "single" deletes an emptied series but keeps the detached event.
		if series, err = c.lockSeries(ctx, seriesID); err != nil {
			return nil, nil, err
		}
	}
	ev, err = domainevent.FindOwnedForUpdate(ctx, c.repo, c.logger, id, userID)
	if err != nil {
		return nil, nil, err
	}
	// The series may have changed before the event was locked: detached by
	// another request, or converted from a one-off. Locking the new series
	// after the event breaks the order, but only in that rare race.
	seriesID, ok := ev.SeriesID()
	if !ok {
		return ev, nil, nil
	}
	if series == nil || series.ID() != seriesID {
		if series, err = c.lockSeries(ctx, seriesID); err != nil {
			return nil, nil, err
		}
	}
	if series == nil {
		// The locked event still refers to the series, so the foreign key
		// keeps the series alive; reaching here means inconsistent data.
		return nil, nil, fmt.Errorf("%w: event %s, series %s", errSeriesMissing, id, seriesID)
	}
	return ev, series, nil
}

// lockSeries returns nil, without an error, for a series that no longer exists.
func (c *UpdateEventCommand) lockSeries(ctx context.Context, id uuid.UUID) (rs.RecurringEventSeries, error) {
	series, err := c.seriesRepo.FindByIDForUpdate(ctx, id)
	if errors.Is(err, rs.ErrRecurringSeriesNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lock series: %w", err)
	}
	return series, nil
}

func (c *UpdateEventCommand) updateOneOff(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput, existing domainevent.Event, rule *rs.RecurrenceRule,
) (UpdateEventOutput, error) {
	if in.Scope == scopeFuture || in.Scope == scopeAll {
		return UpdateEventOutput{}, validationDetail("scope", "NOT_IN_SERIES", "The event does not belong to a series")
	}
	if rule != nil {
		return c.convertToSeries(ctx, userID, in, existing, *rule)
	}
	updated, views, err := c.applyContent(ctx, userID, in, existing)
	if err != nil {
		return UpdateEventOutput{}, err
	}
	if err := c.repo.Update(ctx, updated); err != nil {
		return UpdateEventOutput{}, fmt.Errorf("update event: %w", err)
	}
	return UpdateEventOutput{Event: updated, Tags: views}, nil
}

func (c *UpdateEventCommand) updateOccurrence(
	ctx context.Context,
	userID uuid.UUID,
	in UpdateEventInput,
	existing domainevent.Event,
	series rs.RecurringEventSeries,
	rule *rs.RecurrenceRule,
) (UpdateEventOutput, error) {
	stopsRepeating := in.Recurrence.Present && in.Recurrence.Spec == nil
	switch in.Scope {
	case "":
		return UpdateEventOutput{}, validationDetail("scope", "REQUIRED", "scope is required for an occurrence of a series")
	case scopeSingle:
		// A rule sent back unchanged is not a change, so its range is not
		// checked either.
		if stopsRepeating || (rule != nil && !rule.Equal(series.Rule())) {
			return UpdateEventOutput{}, scopeNotAllowed()
		}
		return c.detach(ctx, userID, in, existing, series)
	case scopeAll:
		if stopsRepeating {
			return UpdateEventOutput{}, scopeNotAllowed()
		}
	}
	return UpdateEventOutput{}, errScopeNotImplemented
}

func scopeNotAllowed() error {
	return validationDetail("scope", "SCOPE_NOT_ALLOWED", "The scope cannot apply this change to the recurrence")
}

// detach turns the occurrence into a one-off event and excludes its date from
// the series, deleting the series when no occurrence is left.
func (c *UpdateEventCommand) detach(
	ctx context.Context,
	userID uuid.UUID,
	in UpdateEventInput,
	existing domainevent.Event,
	series rs.RecurringEventSeries,
) (UpdateEventOutput, error) {
	updated, views, err := c.applyContent(ctx, userID, in, existing)
	if err != nil {
		return UpdateEventOutput{}, err
	}
	updated = updated.DetachFromSeries()
	if err := c.repo.Update(ctx, updated); err != nil {
		return UpdateEventOutput{}, fmt.Errorf("update event: %w", err)
	}

	// The stored start, not the edited one, names the date the pattern
	// would produce, since "single" may move the event to another day.
	excluded, err := series.ExcludeOccurrence(existing.StartAt())
	if err != nil {
		return UpdateEventOutput{}, fmt.Errorf("exclude occurrence: %w", err)
	}
	n, err := c.seriesRepo.CountOccurrences(ctx, series.ID())
	if err != nil {
		return UpdateEventOutput{}, fmt.Errorf("count occurrences: %w", err)
	}
	if n == 0 {
		if err := c.seriesRepo.Delete(ctx, series.ID()); err != nil {
			return UpdateEventOutput{}, fmt.Errorf("delete empty series: %w", err)
		}
	} else if err := c.seriesRepo.Update(ctx, excluded); err != nil {
		return UpdateEventOutput{}, fmt.Errorf("update series: %w", err)
	}
	return UpdateEventOutput{Event: updated, Tags: views}, nil
}

// convertToSeries makes the one-off event the first occurrence of a new
// series, so its ID, share links and subscriptions survive.
func (c *UpdateEventCommand) convertToSeries(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput, existing domainevent.Event, rule rs.RecurrenceRule,
) (UpdateEventOutput, error) {
	if err := checkNewRule(rule); err != nil {
		return UpdateEventOutput{}, err
	}
	occs, err := rs.ExpandOccurrences(in.StartAt, in.EndAt, rule, nil)
	if err != nil {
		return UpdateEventOutput{}, toExpandValidationError(err)
	}
	updated, views, err := c.applyContent(ctx, userID, in, existing)
	if err != nil {
		return UpdateEventOutput{}, err
	}
	series, err := rs.New(uuid.New(), userID, rule, nil)
	if err != nil {
		return UpdateEventOutput{}, fmt.Errorf("create series: %w", err)
	}
	saved, err := c.seriesRepo.Create(ctx, series)
	if err != nil {
		return UpdateEventOutput{}, fmt.Errorf("create series: %w", err)
	}
	updated = updated.WithSeriesID(saved.ID())
	if err := c.repo.Update(ctx, updated); err != nil {
		return UpdateEventOutput{}, fmt.Errorf("update event: %w", err)
	}
	// occs[0] is the event itself; ExpandOccurrences keeps the given instant.
	content := occurrenceContent{title: in.Title, description: in.Description, location: in.Location, url: in.URL}
	rest, err := newOccurrenceEvents(userID, content, saved.ID(), occs[1:], updated.Tags())
	if err != nil {
		return UpdateEventOutput{}, err
	}
	if err := c.seriesRepo.CreateOccurrences(ctx, rest); err != nil {
		return UpdateEventOutput{}, fmt.Errorf("create occurrences: %w", err)
	}
	return UpdateEventOutput{Event: updated, Tags: views, Series: saved}, nil
}

// applyContent returns existing with the requested content and tags.
func (c *UpdateEventCommand) applyContent(
	ctx context.Context, userID uuid.UUID, in UpdateEventInput, existing domainevent.Event,
) (domainevent.Event, []EventTagView, error) {
	newTags, views, err := c.determineTagsInTx(ctx, userID, existing, in.TagIDs)
	if err != nil {
		return nil, nil, err
	}
	updated, err := existing.WithContent(in.Title, in.Description, in.StartAt, in.EndAt, in.Location, in.URL)
	if err != nil {
		return nil, nil, err
	}
	updated, err = updated.WithTags(newTags)
	if err != nil {
		return nil, nil, err
	}
	return updated, views, nil
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

// keepTagsInTx locks the rows of the tags being carried over (FOR SHARE) so
// they cannot be deleted before Update writes them back, and drops any tag
// already deleted since the event was read: writing it back would fail the
// event_tags foreign key. The kept tags and their views keep the original order.
func (c *UpdateEventCommand) keepTagsInTx(
	ctx context.Context, tags []domainevent.EventTag,
) ([]domainevent.EventTag, []EventTagView, error) {
	if len(tags) == 0 {
		return tags, nil, nil
	}
	ids := make([]uuid.UUID, len(tags))
	for i, t := range tags {
		ids[i] = t.TagID()
	}
	locked, err := c.tagRepo.FindByIDsForShare(ctx, ids)
	if err != nil {
		return nil, nil, fmt.Errorf("lock event tags: %w", err)
	}

	byID := make(map[uuid.UUID]tag.Tag, len(locked))
	for _, t := range locked {
		byID[t.ID()] = t
	}
	kept := make([]domainevent.EventTag, 0, len(tags))
	views := make([]EventTagView, 0, len(tags))
	for _, et := range tags {
		t, ok := byID[et.TagID()]
		if !ok {
			continue
		}
		kept = append(kept, et)
		views = append(views, EventTagView{
			ID:    t.ID(),
			Name:  string(t.Name()),
			Color: string(t.Color()),
		})
	}
	return kept, views, nil
}
