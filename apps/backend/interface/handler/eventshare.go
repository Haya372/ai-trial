package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

const errCodeForbidden = "FORBIDDEN"

// CreateShareExecutor is satisfied by eventshareuc.CreateShareCommand.
type CreateShareExecutor interface {
	Execute(
		ctx context.Context, userID uuid.UUID, in eventshareuc.CreateShareInput,
	) (*eventshareuc.CreateShareResult, error)
}

type EventShareHandler struct {
	createShare CreateShareExecutor
	logger      *slog.Logger
}

func NewEventShareHandler(c CreateShareExecutor, logger *slog.Logger) *EventShareHandler {
	return &EventShareHandler{createShare: c, logger: logger}
}

// CreateShare handles POST /events/{id}/shares.
func (h *EventShareHandler) CreateShare(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to POST /events/{id}/shares", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	eventID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid event id")
		return
	}

	in, err := decodeCreateShareInput(r, eventID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	result, err := h.createShare.Execute(r.Context(), u.ID(), in)
	if err != nil {
		h.writeEventShareError(w, r, err)
		return
	}

	body, err := json.Marshal(api.CreateEventShareResponse{
		Url:       "/share/" + result.Token,
		ExpiresAt: result.Share.ExpiresAt(),
	})
	if err != nil {
		h.logger.Error("failed to marshal create share response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(body)
}

func decodeCreateShareInput(r *http.Request, eventID uuid.UUID) (eventshareuc.CreateShareInput, error) {
	in := eventshareuc.CreateShareInput{EventID: eventID}

	var body api.CreateEventShareJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if errors.Is(err, io.EOF) {
			return in, nil
		}
		return eventshareuc.CreateShareInput{}, err
	}
	in.ExpiresAt = body.ExpiresAt
	return in, nil
}

func (h *EventShareHandler) writeEventShareError(w http.ResponseWriter, r *http.Request, err error) {
	writeDomainError(w, r, h.logger, err, map[string]domainErrorResponse{
		domainevent.CodeEventNotFound: {status: http.StatusNotFound, code: errCodeNotFound, message: "Resource not found"},
		domainevent.CodeEventForbidden: {
			status: http.StatusForbidden, code: errCodeForbidden, message: "You do not have access to this resource",
		},
	})
}
