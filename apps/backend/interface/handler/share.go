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
	eventshareuc "github.com/Haya372/ai-trial/backend/usecase/eventshare"
)

const errCodeGone = "GONE"

type GetShareByTokenExecutor interface {
	Execute(ctx context.Context, in eventshareuc.GetShareByTokenInput) (eventshareuc.GetShareByTokenOutput, error)
}

type ShareHandler struct {
	getShareByToken GetShareByTokenExecutor
	logger          *slog.Logger
}

func NewShareHandler(g GetShareByTokenExecutor, logger *slog.Logger) *ShareHandler {
	return &ShareHandler{getShareByToken: g, logger: logger}
}

type shareResponseBody struct {
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	StartAt      time.Time `json:"startAt"`
	EndAt        time.Time `json:"endAt"`
	Location     *string   `json:"location,omitempty"`
	URL          *string   `json:"url,omitempty"`
	IsOwnEvent   bool      `json:"isOwnEvent"`
	IsSubscribed bool      `json:"isSubscribed"`
}

func (h *ShareHandler) GetShareByToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var viewer user.User
	if u, ok := r.Context().Value(ctxkey.User).(user.User); ok && u != nil {
		viewer = u
	}

	out, err := h.getShareByToken.Execute(r.Context(), eventshareuc.GetShareByTokenInput{
		Token:  token,
		Viewer: viewer,
	})
	if err != nil {
		h.writeShareError(w, r, err)
		return
	}

	body := shareResponseBody{
		Title:        out.Title,
		Description:  optionalString(out.Description),
		StartAt:      out.StartAt,
		EndAt:        out.EndAt,
		Location:     optionalString(out.Location),
		URL:          optionalString(out.URL),
		IsOwnEvent:   out.IsOwnEvent,
		IsSubscribed: out.IsSubscribed,
	}

	b, err := json.Marshal(body)
	if err != nil {
		h.logger.Error("failed to marshal share response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// writeShareError maps ErrEventShareNotFound/ErrEventShareExpired to 404/410.
// domainevent.CodeEventNotFound is also mapped to 404: it surfaces only when
// the underlying event is deleted in the narrow window between
// ShareTokenLoader finding the share and loading its event (ON DELETE CASCADE
// makes this race exceedingly rare), and SPEC-004 treats a deleted event the
// same as a missing share, not a server error.
func (h *ShareHandler) writeShareError(w http.ResponseWriter, r *http.Request, err error) {
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
	})
}
