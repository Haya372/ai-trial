package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	domaineventsubscription "github.com/Haya372/ai-trial/backend/domain/eventsubscription"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

const errCodeForbidden = "FORBIDDEN"

// SubscribeToShareExecutor is satisfied by eventsubscriptionuc.SubscribeToShareCommand.
type SubscribeToShareExecutor interface {
	Execute(
		ctx context.Context, in eventsubscriptionuc.SubscribeToShareInput,
	) (eventsubscriptionuc.SubscribeToShareOutput, error)
}

// DeleteSubscriptionExecutor is satisfied by eventsubscriptionuc.DeleteSubscriptionCommand.
// Deliberately separate port from DeleteEventExecutor: different usecase,
// coincidentally same shape.
//
//nolint:iface
type DeleteSubscriptionExecutor interface {
	Execute(ctx context.Context, userID, id uuid.UUID) error
}

type EventSubscriptionHandler struct {
	subscribeToShare   SubscribeToShareExecutor
	deleteSubscription DeleteSubscriptionExecutor
	logger             *slog.Logger
}

func NewEventSubscriptionHandler(
	s SubscribeToShareExecutor, d DeleteSubscriptionExecutor, logger *slog.Logger,
) *EventSubscriptionHandler {
	return &EventSubscriptionHandler{subscribeToShare: s, deleteSubscription: d, logger: logger}
}

type eventSubscriptionResponseBody struct {
	ID        string    `json:"id"`
	EventID   string    `json:"eventId"`
	CreatedAt time.Time `json:"createdAt"`
}

// SubscribeToShare handles POST /shares/{token}/subscriptions.
func (h *EventSubscriptionHandler) SubscribeToShare(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to POST /shares/{token}/subscriptions", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	token := domaineventshare.NewToken(chi.URLParam(r, "token"))

	out, err := h.subscribeToShare.Execute(r.Context(), eventsubscriptionuc.SubscribeToShareInput{
		Token:  token,
		UserID: u.ID(),
	})
	if err != nil {
		h.writeSubscribeError(w, r, err)
		return
	}

	body := eventSubscriptionResponseBody{
		ID:        out.SubscriptionID.String(),
		EventID:   out.EventID.String(),
		CreatedAt: out.CreatedAt,
	}
	b, err := json.Marshal(body)
	if err != nil {
		h.logger.Error("failed to marshal event subscription response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}

	status := http.StatusOK
	if out.Created {
		status = http.StatusCreated
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

// DeleteSubscription handles DELETE /subscriptions/{id}: removing a shared
// event from the authenticated user's own calendar (SPEC-004 "追加の取り消し").
// It only ever deletes the subscription row; the source event and its share
// links are untouched.
func (h *EventSubscriptionHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to DELETE /subscriptions/{id}", "path", r.URL.Path)
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid subscription id")
		return
	}

	if err := h.deleteSubscription.Execute(r.Context(), u.ID(), id); err != nil {
		h.writeDeleteSubscriptionError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeDeleteSubscriptionError maps EVENT_SUBSCRIPTION_NOT_OWNER to 403,
// distinct from the 404 used when the subscription doesn't exist at all
// (Issue #212's acceptance criteria call for the two to be distinguished).
func (h *EventSubscriptionHandler) writeDeleteSubscriptionError(w http.ResponseWriter, r *http.Request, err error) {
	writeDomainError(w, r, h.logger, err, map[string]domainErrorResponse{
		domaineventsubscription.CodeEventSubscriptionNotFound: {
			status: http.StatusNotFound, code: errCodeNotFound, message: msgResourceNotFound,
		},
		domaineventsubscription.CodeNotSubscriptionOwner: {
			status: http.StatusForbidden, code: errCodeForbidden, message: "You cannot delete another user's subscription",
		},
	})
}

// writeSubscribeError maps the same not-found/expired codes as
// ShareHandler.writeShareError (the share token is resolved the same way),
// plus EVENT_SUBSCRIPTION_OWN_EVENT for the "can't add your own event" case.
func (h *EventSubscriptionHandler) writeSubscribeError(w http.ResponseWriter, r *http.Request, err error) {
	writeDomainError(w, r, h.logger, err, map[string]domainErrorResponse{
		domaineventshare.CodeEventShareNotFound: {
			status: http.StatusNotFound, code: errCodeNotFound, message: msgResourceNotFound,
		},
		domaineventshare.CodeEventShareExpired: {
			status: http.StatusGone, code: errCodeGone, message: "Share link has expired",
		},
		domainevent.CodeEventNotFound: {
			status: http.StatusNotFound, code: errCodeNotFound, message: msgResourceNotFound,
		},
		domaineventsubscription.CodeCannotSubscribeToOwnEvent: {
			status: http.StatusForbidden, code: errCodeForbidden, message: "You cannot add your own event to your calendar",
		},
	})
}
