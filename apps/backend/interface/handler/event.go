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

const (
	errCodeNotFound  = "NOT_FOUND"
	errCodeForbidden = "FORBIDDEN"
)

type ListEventsExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.ListEventsInput) ([]eventuc.EventReadModel, error)
}

type UpdateEventExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.UpdateEventInput) (domainevent.Event, error)
}

type EventHandler struct {
	listEvents  ListEventsExecutor
	updateEvent UpdateEventExecutor
	logger      *slog.Logger
}

func NewEventHandler(l ListEventsExecutor, u UpdateEventExecutor, logger *slog.Logger) *EventHandler {
	return &EventHandler{listEvents: l, updateEvent: u, logger: logger}
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

	resp := api.EventsListResponse{
		Events: make([]api.EventResponse, len(events)),
	}
	for i, e := range events {
		ev := api.EventResponse{
			Id:      e.ID,
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
		case domainevent.CodeForbidden:
			response.WriteError(w, http.StatusForbidden, errCodeForbidden, "You do not have permission to access this resource")
		default:
			h.logger.Error("unexpected domain error in event handler", "error", err, "path", r.URL.Path)
			response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		}
	default:
		h.logger.Error("internal error in event handler", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}
