package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	domaineventshare "github.com/Haya372/ai-trial/backend/domain/eventshare"
	"github.com/Haya372/ai-trial/backend/domain/user"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	eventsubscriptionuc "github.com/Haya372/ai-trial/backend/usecase/eventsubscription"
)

// SubscribeToShareExecutor is satisfied by eventsubscriptionuc.SubscribeToShareCommand.
type SubscribeToShareExecutor interface {
	Execute(
		ctx context.Context, in eventsubscriptionuc.SubscribeToShareInput,
	) (eventsubscriptionuc.SubscribeToShareOutput, error)
}

type EventSubscriptionHandler struct {
	subscribeToShare SubscribeToShareExecutor
	logger           *slog.Logger
}

func NewEventSubscriptionHandler(s SubscribeToShareExecutor, logger *slog.Logger) *EventSubscriptionHandler {
	return &EventSubscriptionHandler{subscribeToShare: s, logger: logger}
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

// writeSubscribeError maps the same not-found/expired codes as
// ShareHandler.writeShareError (the share token is resolved the same way),
// plus EVENT_FORBIDDEN for the "can't add your own event" case.
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
		domainevent.CodeEventForbidden: {
			status: http.StatusForbidden, code: errCodeForbidden, message: "You do not have access to this resource",
		},
	})
}
