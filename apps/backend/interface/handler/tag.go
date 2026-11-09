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

// ListTagsExecutor is satisfied by taguc.ListTagsQuery.
type ListTagsExecutor interface {
	Execute(ctx context.Context, userID uuid.UUID) ([]taguc.TagReadModel, error)
}

type TagHandler struct {
	listTags  ListTagsExecutor
	createTag CreateTagExecutor
	logger    *slog.Logger
}

func NewTagHandler(l ListTagsExecutor, c CreateTagExecutor, logger *slog.Logger) *TagHandler {
	return &TagHandler{listTags: l, createTag: c, logger: logger}
}

// ListTags handles GET /tags.
func (h *TagHandler) ListTags(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(ctxkey.User).(user.User)
	if !ok || u == nil {
		h.logger.Warn("unauthorized access to GET /tags", "path", route.Pattern(r))
		response.WriteError(w, http.StatusUnauthorized, errCodeUnauthorized, "Authentication required")
		return
	}

	tags, err := h.listTags.Execute(r.Context(), u.ID())
	if err != nil {
		h.logger.Error("failed to list tags", "error", err, "userID", u.ID())
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}

	resp := api.TagsListResponse{Tags: make([]api.TagResponse, len(tags))}
	for i, tg := range tags {
		resp.Tags[i] = api.TagResponse{
			Id:         tg.ID,
			Name:       tg.Name,
			Color:      api.TagColor(tg.Color),
			EventCount: tg.EventCount,
		}
	}

	out, err := json.Marshal(resp)
	if err != nil {
		h.logger.Error("failed to marshal list tags response", "error", err)
		response.WriteError(w, http.StatusInternalServerError, errCodeInternal, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
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
