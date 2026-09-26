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

	"github.com/Haya372/ai-trial/backend/domain"
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
		case domainevent.CodeEventForbidden:
			response.WriteError(w, http.StatusForbidden, errCodeForbidden, "You do not have access to this resource")
		default:
			h.logger.Error("unexpected domain error in event share handler", "error", err, "path", r.URL.Path)
			response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		}
	default:
		h.logger.Error("internal error in event share handler", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}
