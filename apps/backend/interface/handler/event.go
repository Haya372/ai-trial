package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

const errCodeNotFound = "NOT_FOUND"

// eventResponseBody is the JSON shape for a single event response.
// Defined here because oapi-codegen v2 inlines these fields per-operation.
type eventResponseBody struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	Location    *string   `json:"location,omitempty"`
	URL         *string   `json:"url,omitempty"`
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
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.CreateEventInput) (domainevent.Event, error)
}

type UpdateEventExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.UpdateEventInput) (domainevent.Event, error)
}

type EventHandler struct {
	listEvents  ListEventsExecutor
	createEvent CreateEventExecutor
	updateEvent UpdateEventExecutor
	logger      *slog.Logger
}

func NewEventHandler(
	l ListEventsExecutor, c CreateEventExecutor, u UpdateEventExecutor, logger *slog.Logger,
) *EventHandler {
	return &EventHandler{listEvents: l, createEvent: c, updateEvent: u, logger: logger}
}

func bindDateParam(r *http.Request, name string, dest any) error {
	opts := runtime.BindQueryParameterOptions{Type: "string", Format: "date"}
	return runtime.BindQueryParameterWithOptions("form", true, true, name, r.URL.Query(), dest, opts)
}

func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var params api.GetEventsParams
	if err := bindDateParam(r, "startDate", &params.StartDate); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "invalid startDate parameter")
		return
	}
	if err := bindDateParam(r, "endDate", &params.EndDate); err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "invalid endDate parameter")
		return
	}
	h.GetEvents(w, r, params)
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request, params api.GetEventsParams) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to GET /events", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	in := eventuc.ListEventsInput{
		StartDate: params.StartDate.Time,
		EndDate:   params.EndDate.Add(24 * time.Hour),
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
		ev := eventResponseBody{
			ID:      e.ID,
			Title:   e.Title,
			StartAt: e.StartAt,
			EndAt:   e.EndAt,
		}
		if e.Description != "" {
			ev.Description = &e.Description
		}
		resp.Events[i] = ev
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
	return in
}

func toEventResponse(ev domainevent.Event) api.EventResponse {
	resp := api.EventResponse{
		Id:      ev.ID(),
		Title:   ev.Title(),
		StartAt: ev.StartAt(),
		EndAt:   ev.EndAt(),
	}
	if ev.Description() != "" {
		d := ev.Description()
		resp.Description = &d
	}
	if ev.Location() != "" {
		l := ev.Location()
		resp.Location = &l
	}
	if ev.URL() != "" {
		eu := ev.URL()
		resp.Url = &eu
	}
	return resp
}

// CreateEvent handles POST /events.
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to POST /events", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	var body api.CreateEventJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Warn("invalid JSON body for POST /events", "error", err)
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	ev, err := h.createEvent.Execute(r.Context(), u.ID(), buildCreateEventInput(body))
	if err != nil {
		h.writeEventError(w, r, err)
		return
	}

	out, err := json.Marshal(toEventResponse(ev))
	if err != nil {
		h.logger.Error("failed to marshal create event response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to PUT /events/{id}", "path", r.URL.Path)
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

	ev, err := h.updateEvent.Execute(r.Context(), u.ID(), in)
	if err != nil {
		h.writeEventError(w, r, err)
		return
	}

	respBody, err := json.Marshal(toEventResponse(ev))
	if err != nil {
		h.logger.Error("failed to marshal update event response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBody)
}

func decodeUpdateEventInput(r *http.Request, id uuid.UUID) (eventuc.UpdateEventInput, error) {
	var body api.UpdateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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
	return in, nil
}

func (h *EventHandler) writeEventError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *domain.ValidationError
	var de *domain.DomainError

	switch {
	case errors.As(err, &ve):
		details := make([]response.ErrorDetail, len(ve.Details))
		for i, d := range ve.Details {
			details[i] = response.ErrorDetail{Field: d.Field, Code: d.Code, Message: d.Message}
		}
		response.WriteValidationError(w, details)
	case errors.As(err, &de):
		switch de.Code() {
		case domainevent.CodeEventNotFound:
			response.WriteError(w, http.StatusNotFound, errCodeNotFound, "Resource not found")
		default:
			h.logger.Error("unexpected domain error in event handler", "error", err, "path", r.URL.Path)
			response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		}
	default:
		h.logger.Error("internal error in event handler", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}
