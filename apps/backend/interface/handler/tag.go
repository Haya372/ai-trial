package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/domain/user"
	api "github.com/Haya372/ai-trial/backend/interface/api/generated"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler/response"
	"github.com/Haya372/ai-trial/backend/interface/route"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

// CreateTagExecutor is satisfied by taguc.CreateTagCommand.
type CreateTagExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID, in taguc.CreateTagInput) (domaintag.Tag, error)
}

type TagHandler struct {
	createTag CreateTagExecutor
	logger    *slog.Logger
}

func NewTagHandler(c CreateTagExecutor, logger *slog.Logger) *TagHandler {
	return &TagHandler{createTag: c, logger: logger}
}

// CreateTag handles POST /tags.
func (h *TagHandler) CreateTag(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to POST /tags", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	var body api.CreateTagJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.logger.Warn("invalid JSON body for POST /tags", "error", err)
		response.WriteError(w, http.StatusBadRequest, errCodeValidation, "Invalid request body")
		return
	}

	tg, err := h.createTag.Execute(r.Context(), u.ID(), taguc.CreateTagInput{
		Name:  body.Name,
		Color: string(body.Color),
	})
	if err != nil {
		h.writeTagError(w, r, err)
		return
	}

	// A tag that was just created has no events attached yet.
	out, err := json.Marshal(api.TagResponse{
		Id:         tg.ID(),
		Name:       string(tg.Name()),
		Color:      api.TagColor(tg.Color()),
		EventCount: 0,
	})
	if err != nil {
		h.logger.Error("failed to marshal create tag response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(out)
}

// writeTagError reports a duplicate name and the tag limit with the same 409
// and no details; the client tells them apart by refetching GET /tags (SPEC-008).
func (h *TagHandler) writeTagError(w http.ResponseWriter, r *http.Request, err error) {
	writeDomainError(w, r, h.logger, err, map[string]domainErrorResponse{
		domaintag.CodeTagNameTaken: {
			status: http.StatusConflict, code: errCodeConflict, message: "Tag name already exists",
		},
		domaintag.CodeTagLimitExceeded: {
			status: http.StatusConflict, code: errCodeConflict, message: "Tag limit exceeded",
		},
	})
}
