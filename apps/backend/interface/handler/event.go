package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime"

	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

type ListEventsExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in eventuc.ListEventsInput) ([]eventuc.EventReadModel, error)
}

type EventHandler struct {
	listEvents ListEventsExecutor
	logger     *slog.Logger
}

func NewEventHandler(l ListEventsExecutor, logger *slog.Logger) *EventHandler {
	return &EventHandler{listEvents: l, logger: logger}
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
