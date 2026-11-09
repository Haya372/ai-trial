package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

const (
	fieldTagName      = "name"
	fieldTagColor     = "color"
	codeInvalidFormat = "INVALID_FORMAT"
	testTagName       = "Work"
)

type stubCreateTagExec struct {
	fn func(ctx context.Context, userID uuid.UUID, in taguc.CreateTagInput) (domaintag.Tag, error)
}

func (s *stubCreateTagExec) Execute(
	ctx context.Context, userID uuid.UUID, in taguc.CreateTagInput,
) (domaintag.Tag, error) {
	if s.fn == nil {
		return nil, errUnexpected
	}
	return s.fn(ctx, userID, in)
}

func createTagRequest(t *testing.T, body string, userID *uuid.UUID) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "/tags", bytes.NewReader([]byte(body)))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if userID != nil {
		req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(*userID, "u@ex.com", "U")))
	}
	return req
}

func decodeErrorBody(t *testing.T, w *httptest.ResponseRecorder) (string, []map[string]any) {
	t.Helper()
	var resp struct {
		Code    string           `json:"code"`
		Details []map[string]any `json:"details"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return resp.Code, resp.Details
}

func TestTagHandler_CreateTag_Success_Returns201(t *testing.T) {
	userID := uuid.New()
	tagID := uuid.New()

	stub := &stubCreateTagExec{
		fn: func(_ context.Context, uid uuid.UUID, in taguc.CreateTagInput) (domaintag.Tag, error) {
			if uid != userID {
				t.Errorf("userID: got %v, want %v", uid, userID)
			}
			if in.Name != " Work " || in.Color != "teal" {
				t.Errorf("input: got %+v", in)
			}
			return domaintag.New(tagID, userID, domaintag.TagName(testTagName), domaintag.ColorTeal), nil
		},
	}
	h := handler.NewTagHandler(&stubListTagsExec{}, stub, testLogger)
	w := httptest.NewRecorder()

	h.CreateTag(w, createTagRequest(t, `{"name":" Work ","color":"teal"}`, &userID))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q", ct)
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := map[string]any{
		"id": tagID.String(), fieldTagName: testTagName, fieldTagColor: "teal", "eventCount": float64(0),
	}
	for k, v := range want {
		if resp[k] != v {
			t.Errorf("%s: got %v, want %v", k, resp[k], v)
		}
	}
}

func TestTagHandler_CreateTag_Unauthorized_Returns401(t *testing.T) {
	h := handler.NewTagHandler(&stubListTagsExec{}, &stubCreateTagExec{}, testLogger)
	w := httptest.NewRecorder()

	h.CreateTag(w, createTagRequest(t, `{"name":"work","color":"red"}`, nil))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTagHandler_CreateTag_InvalidJSON_Returns400WithoutDetails(t *testing.T) {
	userID := uuid.New()
	h := handler.NewTagHandler(&stubListTagsExec{}, &stubCreateTagExec{}, testLogger)
	w := httptest.NewRecorder()

	h.CreateTag(w, createTagRequest(t, `{"name":`, &userID))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	code, details := decodeErrorBody(t, w)
	if code != codeValidationError {
		t.Errorf("code: got %q, want %q", code, codeValidationError)
	}
	if len(details) != 0 {
		t.Errorf("expected no details, got %v", details)
	}
}

func TestTagHandler_CreateTag_ValidationError_Returns400WithDetails(t *testing.T) {
	userID := uuid.New()
	stub := &stubCreateTagExec{
		fn: func(context.Context, uuid.UUID, taguc.CreateTagInput) (domaintag.Tag, error) {
			return nil, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: fieldTagColor, Code: codeInvalidFormat, Message: "Tag color must be one of the preset colors"},
			}}
		},
	}
	h := handler.NewTagHandler(&stubListTagsExec{}, stub, testLogger)
	w := httptest.NewRecorder()

	h.CreateTag(w, createTagRequest(t, `{"name":"work","color":"black"}`, &userID))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	code, details := decodeErrorBody(t, w)
	if code != codeValidationError {
		t.Errorf("code: got %q", code)
	}
	if len(details) != 1 || details[0]["field"] != fieldTagColor || details[0]["code"] != codeInvalidFormat {
		t.Errorf("details: got %v", details)
	}
}

func TestTagHandler_CreateTag_Conflict_Returns409WithoutDetails(t *testing.T) {
	for _, domainErr := range []error{domaintag.ErrTagNameTaken, domaintag.ErrTagLimitExceeded} {
		t.Run(domainErr.Error(), func(t *testing.T) {
			userID := uuid.New()
			stub := &stubCreateTagExec{
				fn: func(context.Context, uuid.UUID, taguc.CreateTagInput) (domaintag.Tag, error) {
					return nil, domainErr
				},
			}
			h := handler.NewTagHandler(&stubListTagsExec{}, stub, testLogger)
			w := httptest.NewRecorder()

			h.CreateTag(w, createTagRequest(t, `{"name":"work","color":"red"}`, &userID))

			if w.Code != http.StatusConflict {
				t.Fatalf("expected 409, got %d", w.Code)
			}
			code, details := decodeErrorBody(t, w)
			if code != "CONFLICT" {
				t.Errorf("code: got %q, want CONFLICT", code)
			}
			if len(details) != 0 {
				t.Errorf("expected no details, got %v", details)
			}
		})
	}
}

func TestTagHandler_CreateTag_InternalError_Returns500(t *testing.T) {
	userID := uuid.New()
	stub := &stubCreateTagExec{
		fn: func(context.Context, uuid.UUID, taguc.CreateTagInput) (domaintag.Tag, error) {
			return nil, errInternal
		},
	}
	logger, logs := newBufferLogger()
	h := handler.NewTagHandler(&stubListTagsExec{}, stub, logger)
	w := httptest.NewRecorder()

	h.CreateTag(w, createTagRequest(t, `{"name":"work","color":"red"}`, &userID))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(logs.String(), errInternal.Error()) {
		t.Errorf("expected the internal error to be logged, got:\n%s", logs.String())
	}
}

type stubListTagsExec struct {
	fn func(ctx context.Context, userID uuid.UUID) ([]taguc.TagReadModel, error)
}

func (s *stubListTagsExec) Execute(ctx context.Context, userID uuid.UUID) ([]taguc.TagReadModel, error) {
	if s.fn == nil {
		return nil, errUnexpected
	}
	return s.fn(ctx, userID)
}

func listTagsRequest(t *testing.T, userID *uuid.UUID) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/tags", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if userID != nil {
		req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(*userID, "u@ex.com", "U")))
	}
	return req
}

type tagsListBody struct {
	Tags []struct {
		ID         uuid.UUID `json:"id"`
		Name       string    `json:"name"`
		Color      string    `json:"color"`
		EventCount *int      `json:"eventCount"`
	} `json:"tags"`
}

func TestTagHandler_ListTags_Success_Returns200InGivenOrder(t *testing.T) {
	userID := uuid.New()
	first, second := uuid.New(), uuid.New()
	stub := &stubListTagsExec{
		fn: func(_ context.Context, uid uuid.UUID) ([]taguc.TagReadModel, error) {
			if uid != userID {
				t.Errorf("userID: got %v, want %v", uid, userID)
			}
			return []taguc.TagReadModel{
				{ID: first, Name: testTagName, Color: "green", EventCount: 3},
				{ID: second, Name: "Private", Color: "red", EventCount: 0},
			}, nil
		},
	}
	h := handler.NewTagHandler(stub, &stubCreateTagExec{}, testLogger)
	w := httptest.NewRecorder()

	h.ListTags(w, listTagsRequest(t, &userID))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q", ct)
	}
	var body tagsListBody
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(body.Tags))
	}
	want := []struct {
		id         uuid.UUID
		name       string
		color      string
		eventCount int
	}{
		{first, testTagName, "green", 3},
		{second, "Private", "red", 0},
	}
	for i, wt := range want {
		got := body.Tags[i]
		if got.ID != wt.id || got.Name != wt.name || got.Color != wt.color ||
			got.EventCount == nil || *got.EventCount != wt.eventCount {
			t.Errorf("tags[%d]: got %+v, want %+v", i, got, wt)
		}
	}
}

func TestTagHandler_ListTags_NoTags_ReturnsEmptyArray(t *testing.T) {
	userID := uuid.New()
	stub := &stubListTagsExec{
		fn: func(context.Context, uuid.UUID) ([]taguc.TagReadModel, error) {
			return nil, nil
		},
	}
	h := handler.NewTagHandler(stub, &stubCreateTagExec{}, testLogger)
	w := httptest.NewRecorder()

	h.ListTags(w, listTagsRequest(t, &userID))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got := strings.TrimSpace(w.Body.String()); got != `{"tags":[]}` {
		t.Errorf("body: got %s, want {\"tags\":[]}", got)
	}
}

func TestTagHandler_ListTags_Unauthorized_Returns401(t *testing.T) {
	h := handler.NewTagHandler(&stubListTagsExec{}, &stubCreateTagExec{}, testLogger)
	w := httptest.NewRecorder()

	h.ListTags(w, listTagsRequest(t, nil))

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTagHandler_ListTags_InternalError_Returns500(t *testing.T) {
	userID := uuid.New()
	stub := &stubListTagsExec{
		fn: func(context.Context, uuid.UUID) ([]taguc.TagReadModel, error) {
			return nil, errInternal
		},
	}
	logger, logs := newBufferLogger()
	h := handler.NewTagHandler(stub, &stubCreateTagExec{}, logger)
	w := httptest.NewRecorder()

	h.ListTags(w, listTagsRequest(t, &userID))

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if !strings.Contains(logs.String(), errInternal.Error()) {
		t.Errorf("expected the internal error to be logged, got:\n%s", logs.String())
	}
}
