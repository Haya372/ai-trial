package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Haya372/ai-trial/backend/domain"
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

func (h *ShareHandler) writeShareError(w http.ResponseWriter, r *http.Request, err error) {
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
		case domaineventshare.CodeEventShareNotFound:
			response.WriteError(w, http.StatusNotFound, errCodeNotFound, "Resource not found")
		case domaineventshare.CodeEventShareExpired:
			response.WriteError(w, http.StatusGone, errCodeGone, "Share link has expired")
		default:
			h.logger.Error("unexpected domain error in share handler", "error", err, "path", r.URL.Path)
			response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		}
	default:
		h.logger.Error("internal error in share handler", "error", err, "path", r.URL.Path)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
	}
}
