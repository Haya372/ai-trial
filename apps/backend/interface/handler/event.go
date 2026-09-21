package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime"

	"github.com/Haya372/ai-trial/backend/domain"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

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
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.CreateEventInput) (eventuc.EventDetail, error)
}

type EventHandler struct {
	listEvents  ListEventsExecutor
	createEvent CreateEventExecutor
	logger      *slog.Logger
}

func NewEventHandler(l ListEventsExecutor, c CreateEventExecutor, logger *slog.Logger) *EventHandler {
	return &EventHandler{listEvents: l, createEvent: c, logger: logger}
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

// toEventResponseBody maps a usecase EventDetail to the JSON response shape.
func toEventResponseBody(d eventuc.EventDetail) eventResponseBody {
	resp := eventResponseBody{
		ID:      d.ID,
		Title:   d.Title,
		StartAt: d.StartAt,
		EndAt:   d.EndAt,
	}
	if d.Description != "" {
		resp.Description = &d.Description
	}
	if d.Location != "" {
		resp.Location = &d.Location
	}
	if d.URL != "" {
		resp.URL = &d.URL
	}
	return resp
}

// writeCreateEventValidationError converts a domain.ValidationError to a 400 response.
func writeCreateEventValidationError(w http.ResponseWriter, ve *domain.ValidationError) {
	details := make([]response.ErrorDetail, len(ve.Details))
	for i, d := range ve.Details {
		details[i] = response.ErrorDetail{Field: d.Field, Code: d.Code, Message: d.Message}
	}
	response.WriteValidationError(w, details)
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

	detail, err := h.createEvent.Execute(r.Context(), u.ID(), buildCreateEventInput(body))
	if err != nil {
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			writeCreateEventValidationError(w, ve)
			return
		}
		h.logger.Error("failed to create event", "error", err, "userID", u.ID())
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}

	out, err := json.Marshal(toEventResponseBody(detail))
	if err != nil {
		h.logger.Error("failed to marshal create event response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}
