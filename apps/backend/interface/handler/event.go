package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	"github.com/Haya372/ai-trial/backend/interface/route"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const errCodeNotFound = "NOT_FOUND"

// eventTagResponseBody is the JSON shape for a single tag embedded in event responses.
type eventTagResponseBody struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// eventResponseBody is the JSON shape for a single event response.
// Defined here because oapi-codegen v2 inlines these fields per-operation.
type eventResponseBody struct {
	ID              uuid.UUID                    `json:"id"`
	Title           string                       `json:"title"`
	Description     *string                      `json:"description,omitempty"`
	StartAt         time.Time                    `json:"startAt"`
	EndAt           time.Time                    `json:"endAt"`
	Location        *string                      `json:"location,omitempty"`
	URL             *string                      `json:"url,omitempty"`
	IsSubscribed    bool                         `json:"isSubscribed"`
	SubscriptionID  *uuid.UUID                   `json:"subscriptionId"`
	Tags            []eventTagResponseBody       `json:"tags"`
	RecurringSeries *api.RecurringSeriesResponse `json:"recurringSeries"`
}

// eventsListResponseBody is the JSON shape for the list events response.
type eventsListResponseBody struct {
	Events []eventResponseBody `json:"events"`
}

// ListEventsExecutor is satisfied by eventuc.ListEventsCommand.
type ListEventsExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.ListEventsInput) ([]eventuc.EventReadModel, error)
}

// CreateEventExecutor is satisfied by eventuc.CreateEventCommand.
type CreateEventExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error)
}

type UpdateEventExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error)
}

// DeleteEventExecutor is a port deliberately kept separate from
// DeleteSubscriptionExecutor: different usecase, coincidentally same shape.
//
//nolint:iface // separate ports for separate usecases, see above
type DeleteEventExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, id uuid.UUID) error
}

type EventHandler struct {
	listEvents  ListEventsExecutor
	createEvent CreateEventExecutor
	updateEvent UpdateEventExecutor
	deleteEvent DeleteEventExecutor
	logger      *slog.Logger
}

func NewEventHandler(
	l ListEventsExecutor, c CreateEventExecutor, u UpdateEventExecutor, d DeleteEventExecutor, logger *slog.Logger,
) *EventHandler {
	return &EventHandler{listEvents: l, createEvent: c, updateEvent: u, deleteEvent: d, logger: logger}
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params api.GetEventsParams
	var err error
	if params.StartAt, err = parseUTCDateTime(r.URL.Query().Get("startAt")); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "invalid startAt parameter")
		return
	}
	if params.EndAt, err = parseUTCDateTime(r.URL.Query().Get("endAt")); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "invalid endAt parameter")
		return
	}
	if !params.EndAt.After(params.StartAt) {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "endAt must be after startAt")
		return
	}
	if tagIDStr := r.URL.Query().Get("tagId"); tagIDStr != "" {
		tagUUID, err := uuid.Parse(tagIDStr)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, errCodeValidation, "invalid tagId parameter")
			return
		}
		params.TagId = &tagUUID
	}
	h.GetEvents(w, r, params)
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request, params api.GetEventsParams) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to GET /events", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	in := eventuc.ListEventsInput{
		StartDate: params.StartAt,
		EndDate:   params.EndAt,
		TagID:     params.TagId,
	}
	events, err := h.listEvents.Execute(r.Context(), u.ID(), in)
	if err != nil {
		h.logger.Error("failed to list events", "error", err, "userID", u.ID())
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}

	resp := eventsListResponseBody{
		Events: make([]eventResponseBody, len(events)),
	}
	for i, e := range events {
		tags := make([]eventTagResponseBody, len(e.Tags))
		for j, t := range e.Tags {
			tags[j] = eventTagResponseBody{ID: t.ID, Name: t.Name, Color: t.Color}
		}
		resp.Events[i] = eventResponseBody{
			ID:              e.ID,
			Title:           e.Title,
			StartAt:         e.StartAt,
			EndAt:           e.EndAt,
			Description:     optionalString(e.Description),
			Location:        optionalString(e.Location),
			URL:             optionalString(e.URL),
			IsSubscribed:    e.IsSubscribed,
			SubscriptionID:  e.SubscriptionID,
			Tags:            tags,
			RecurringSeries: toRecurringSeriesResponseFromReadModel(e.RecurringSeries),
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("failed to marshal events response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

// buildCreateEventInput converts the API request body to a usecase input.
func buildCreateEventInput(body api.CreateEventJSONRequestBody) eventuc.CreateEventInput {
	in := eventuc.CreateEventInput{
		Title:   body.Title,
		StartAt: body.StartAt,
		EndAt:   body.EndAt,
	}
	if body.Description != nil {
		in.Description = *body.Description
	}
	if body.Location != nil {
		in.Location = *body.Location
	}
	if body.Url != nil {
		in.URL = *body.Url
	}
	if body.TagIds != nil {
		in.TagIDs = make([]uuid.UUID, len(*body.TagIds))
		copy(in.TagIDs, *body.TagIds)
	}
	in.Recurrence = toRecurrenceInput(body.Recurrence)
	return in
}

// toRecurrenceInput copies the recurrence fields without validating them.
// Returns nil when spec is nil.
func toRecurrenceInput(spec *api.RecurrenceSpec) *eventuc.RecurrenceInput {
	if spec == nil {
		return nil
	}
	ri := &eventuc.RecurrenceInput{
		PatternType: string(spec.Pattern.Type),
		EndType:     string(spec.End.Type),
		UntilDate:   spec.End.UntilDate,
		Count:       spec.End.Count,
		Timezone:    spec.Timezone,
	}
	if spec.Pattern.MonthlyNth != nil {
		s := string(*spec.Pattern.MonthlyNth)
		ri.MonthlyNth = &s
	}
	return ri
}

// toEventResponse builds the EventResponse for create/update endpoints.
// These endpoints only act on the caller's own event, so IsSubscribed/SubscriptionId
// are always false/nil.
func toEventResponse(
	ev domainevent.Event, tagViews []eventuc.EventTagView, series rs.RecurringEventSeries,
) api.EventResponse {
	apiTags := make([]api.EventTagResponse, len(tagViews))
	for i, t := range tagViews {
		apiTags[i] = api.EventTagResponse{
			Id:    t.ID,
			Name:  t.Name,
			Color: api.TagColor(t.Color),
		}
	}
	return api.EventResponse{
		Id:              ev.ID(),
		Title:           ev.Title(),
		StartAt:         ev.StartAt(),
		EndAt:           ev.EndAt(),
		Description:     optionalString(ev.Description()),
		Location:        optionalString(ev.Location()),
		Url:             optionalString(ev.URL()),
		IsSubscribed:    false,
		SubscriptionId:  nil,
		Tags:            apiTags,
		RecurringSeries: toRecurringSeriesResponse(series),
	}
}

// toRecurringSeriesResponse converts a domain RecurringEventSeries to the API
// response type. Returns nil when series is nil. It goes through the read
// model so POST/PUT and GET /events build the response in one place.
func toRecurringSeriesResponse(series rs.RecurringEventSeries) *api.RecurringSeriesResponse {
	if series == nil {
		return nil
	}
	rule := series.Rule()
	m := &eventuc.RecurringSeriesReadModel{
		ID:          series.ID(),
		PatternType: string(rule.Pattern.Type()),
		EndType:     string(rule.End.Type()),
		Timezone:    rule.Timezone.Name(),
	}
	if nth, ok := rule.Pattern.MonthlyNth(); ok {
		s := string(nth)
		m.MonthlyNth = &s
	}
	if ud, ok := rule.End.UntilDate(); ok {
		s := ud.String()
		m.UntilDate = &s
	}
	if c, ok := rule.End.Count(); ok {
		m.Count = &c
	}
	return toRecurringSeriesResponseFromReadModel(m)
}

// toRecurringSeriesResponseFromReadModel is the single place that builds the
// recurringSeries response for GET /events, POST /events and PUT /events/{id}
// (the latter two via toRecurringSeriesResponse), so changing it changes every
// endpoint's response. Returns nil when m is nil.
func toRecurringSeriesResponseFromReadModel(m *eventuc.RecurringSeriesReadModel) *api.RecurringSeriesResponse {
	if m == nil {
		return nil
	}
	var apiNth *api.RecurrencePatternMonthlyNth
	if m.MonthlyNth != nil {
		v := api.RecurrencePatternMonthlyNth(*m.MonthlyNth)
		apiNth = &v
	}
	return &api.RecurringSeriesResponse{
		Id: m.ID,
		Pattern: api.RecurrencePattern{
			Type:       api.RecurrencePatternType(m.PatternType),
			MonthlyNth: apiNth,
		},
		End: api.RecurrenceEnd{
			Type:      api.RecurrenceEndType(m.EndType),
			UntilDate: m.UntilDate,
			Count:     m.Count,
		},
		Timezone: m.Timezone,
	}
}

// optionalString returns nil for an empty string, so callers get the same
// omitempty behavior GetEvents and toEventResponse both need for optional fields.
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateEvent handles POST /events.
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to POST /events", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	var body api.CreateEventJSONRequestBody
	if err := decodeEventBody(r, &body); err != nil {
		h.logger.Warn("invalid JSON body for POST /events", "error", err)
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	out, err := h.createEvent.Execute(r.Context(), u.ID(), buildCreateEventInput(body))
	if err != nil {
		h.writeEventError(w, r, err)
		return
	}

	respBody, err := json.Marshal(toEventResponse(out.Event, out.Tags, out.Series))
	if err != nil {
		h.logger.Error("failed to marshal create event response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(respBody)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to PUT /events/{id}", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid event id")
		return
	}

	in, err := decodeUpdateEventInput(r, id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	out, err := h.updateEvent.Execute(r.Context(), u.ID(), in)
	if err != nil {
		h.writeEventError(w, r, err)
		return
	}

	respBody, err := json.Marshal(toEventResponse(out.Event, out.Tags, out.Series))
	if err != nil {
		h.logger.Error("failed to marshal update event response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBody)
}

// DeleteEvent handles DELETE /events/{id}.
func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to DELETE /events/{id}", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid event id")
		return
	}

	if err := h.deleteEvent.Execute(r.Context(), u.ID(), id); err != nil {
		h.writeEventError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// eventBodyDateTimes mirrors the datetime fields of the create/update request bodies
// as strings so that requireUTCDateTimes can see their original notation.
type eventBodyDateTimes struct {
	StartAt *string `json:"startAt"`
	EndAt   *string `json:"endAt"`
}

func decodeEventBody(r *http.Request, dst any) error {
	_, err := decodeEventBodyRaw(r, dst)
	return err
}

// decodeEventBodyRaw is decodeEventBody that also returns the raw body, for
// a caller that needs to see more of the original JSON.
func decodeEventBodyRaw(r *http.Request, dst any) ([]byte, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return nil, err
	}
	var dt eventBodyDateTimes
	if err := json.Unmarshal(raw, &dt); err != nil {
		return nil, err
	}
	return raw, requireUTCDateTimes(dt.StartAt, dt.EndAt)
}

// updateBodyRecurrence keeps recurrence as raw JSON: the generated
// *RecurrenceSpec is nil both when the key is omitted (keep the recurrence)
// and when it is null (stop repeating), while RawMessage stays nil only for
// an omitted key.
type updateBodyRecurrence struct {
	Recurrence json.RawMessage `json:"recurrence"`
}

func decodeUpdateEventInput(r *http.Request, id uuid.UUID) (eventuc.UpdateEventInput, error) {
	var body api.UpdateEventRequest
	raw, err := decodeEventBodyRaw(r, &body)
	if err != nil {
		return eventuc.UpdateEventInput{}, err
	}
	var rec updateBodyRecurrence
	if err := json.Unmarshal(raw, &rec); err != nil {
		return eventuc.UpdateEventInput{}, err
	}

	in := eventuc.UpdateEventInput{
		ID:      id,
		Title:   body.Title,
		StartAt: body.StartAt,
		EndAt:   body.EndAt,
	}
	if body.Description != nil {
		in.Description = *body.Description
	}
	if body.Location != nil {
		in.Location = *body.Location
	}
	if body.Url != nil {
		in.URL = *body.Url
	}
	// nil tagIds = keep existing; non-nil (even empty) = replace
	if body.TagIds != nil {
		uuids := make([]uuid.UUID, len(*body.TagIds))
		copy(uuids, *body.TagIds)
		in.TagIDs = &uuids
	}
	in.Recurrence = eventuc.RecurrenceUpdate{
		Present: rec.Recurrence != nil,
		Spec:    toRecurrenceInput(body.Recurrence),
	}
	if body.Scope != nil {
		in.Scope = string(*body.Scope)
	}
	return in, nil
}

func (h *EventHandler) writeEventError(w http.ResponseWriter, r *http.Request, err error) {
	// The domain error has no field, and the mapping below cannot return
	// details, so name startAt here for the form to show it inline.
	if errors.Is(err, rs.ErrRecurringOccurrenceDateChangeNotAllowed) {
		response.WriteValidationError(w, []response.ErrorDetail{{
			Field:   "startAt",
			Code:    rs.CodeRecurringOccurrenceDateChangeNotAllowed,
			Message: "The date cannot change without changing the recurrence",
		}})
		return
	}
	writeDomainError(w, r, h.logger, err, map[string]domainErrorResponse{
		domainevent.CodeEventNotFound: {status: http.StatusNotFound, code: errCodeNotFound, message: msgResourceNotFound},
	})
}
