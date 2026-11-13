//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	domaintag "github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	"github.com/Haya372/ai-trial/backend/infrastructure/repository"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	mw "github.com/Haya372/ai-trial/backend/interface/middleware"
	"github.com/Haya372/ai-trial/backend/usecase"
	authuc "github.com/Haya372/ai-trial/backend/usecase/auth"
	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

func buildTagTestRouter() *chi.Mux {
	logger := slog.New(slog.DiscardHandler)
	userRepo := repository.NewUserRepository(routeTestPool, testTracerProvider)
	sessRepo := repository.NewSessionRepository(routeTestPool, testTracerProvider)
	tagRepo := repository.NewTagRepository(routeTestPool, logger, testTracerProvider)
	txMgr := db.NewPgxTxManager(routeTestPool)

	signup := authuc.NewSignupCommand(userRepo, sessRepo, txMgr, usecase.RealClock{})
	login := authuc.NewLoginCommand(userRepo, sessRepo, usecase.RealClock{})
	logout := authuc.NewLogoutCommand(sessRepo)

	auth := handler.NewAuthHandler(signup, login, logout, logger)
	tg := handler.NewTagHandler(
		taguc.NewListTagsQuery(repository.NewTagQueryRepository(routeTestPool, logger, testTracerProvider)),
		taguc.NewCreateTagCommand(tagRepo, txMgr),
		taguc.NewUpdateTagCommand(tagRepo, txMgr),
		taguc.NewDeleteTagCommand(tagRepo, txMgr),
		logger,
	)

	r := chi.NewRouter()
	r.Post("/auth/signup", auth.Signup)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/auth/me", auth.GetMe)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Get("/tags", tg.ListTags)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Post("/tags", tg.CreateTag)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Put("/tags/{id}", tg.UpdateTag)
	r.With(mw.RequireAuth(sessRepo, userRepo, logger)).Delete("/tags/{id}", tg.DeleteTag)
	return r
}

func postTag(t *testing.T, router *chi.Mux, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/tags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func insertRouteTestTags(t *testing.T, userID string, n int) {
	t.Helper()
	for i := range n {
		if _, err := routeTestPool.Exec(context.Background(),
			"INSERT INTO tags (user_id, name, color) VALUES ($1, $2, 'red')",
			userID, fmt.Sprintf("tag-%02d", i),
		); err != nil {
			t.Fatalf("insert tag: %v", err)
		}
	}
}

func tagRowCount(t *testing.T, userID string) int {
	t.Helper()
	var count int
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM tags WHERE user_id = $1", userID,
	).Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	return count
}

func TestRoute_CreateTag_unauthenticated_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()

	rec := postTag(t, router, nil, `{"name":"work","color":"red"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateTag_success_returns201AndStoresTrimmedName(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-create@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	rec := postTag(t, router, cookie, `{"name":"  Work  ","color":"indigo"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Color      string `json:"color"`
		EventCount *int   `json:"eventCount"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Name != testTagName || body.Color != "indigo" {
		t.Errorf("got name=%q color=%q, want Work/indigo", body.Name, body.Color)
	}
	if body.EventCount == nil || *body.EventCount != 0 {
		t.Errorf("eventCount: got %v, want 0", body.EventCount)
	}

	var storedUserID, storedName, storedColor string
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT user_id, name, color FROM tags WHERE id = $1", body.ID,
	).Scan(&storedUserID, &storedName, &storedColor); err != nil {
		t.Fatalf("select stored tag: %v", err)
	}
	if storedUserID != userID || storedName != testTagName || storedColor != "indigo" {
		t.Errorf("stored row: got (%s, %q, %q), want (%s, Work, indigo)", storedUserID, storedName, storedColor, userID)
	}
}

func TestRoute_CreateTag_invalidInput_returns400WithDetails(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-invalid@ex.com")

	rec := postTag(t, router, cookie, `{"name":"   ","color":"black"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Code    string `json:"code"`
		Details []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"details"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Code != codeValidationError || len(body.Details) != 2 {
		t.Fatalf("got %+v, want VALIDATION_ERROR with 2 details", body)
	}
	if body.Details[0].Field != fieldTagName || body.Details[0].Code != "REQUIRED" ||
		body.Details[1].Field != fieldTagColor || body.Details[1].Code != codeInvalidFormat {
		t.Errorf("details: got %+v", body.Details)
	}
}

func TestRoute_CreateTag_duplicateNameIgnoringCase_returns409(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-dup@ex.com")

	if rec := postTag(t, router, cookie, `{"name":"Work","color":"red"}`); rec.Code != http.StatusCreated {
		t.Fatalf("first create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	rec := postTag(t, router, cookie, `{"name":"WORK","color":"blue"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if code, _ := decodeErrorBody(t, rec); code != "TAG_NAME_TAKEN" {
		t.Errorf("code: got %q, want TAG_NAME_TAKEN", code)
	}
}

func TestRoute_CreateTag_sameNameAsAnotherUsersTag_returns201(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	ownerCookie := signupAndGetCookie(t, router, "tag-owner@ex.com")
	otherCookie := signupAndGetCookie(t, router, "tag-other@ex.com")
	otherID := getUserIDFromCookie(t, router, otherCookie)
	insertRouteTestTags(t, otherID, domaintag.MaxTagsPerUser)

	if rec := postTag(t, router, otherCookie, `{"name":"Work","color":"red"}`); rec.Code != http.StatusConflict {
		t.Fatalf("other user at the limit: expected 409, got %d", rec.Code)
	}
	rec := postTag(t, router, ownerCookie, `{"name":"tag-00","color":"red"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 since names and the limit are per user, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_CreateTag_atLimit_returns409(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-limit@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	insertRouteTestTags(t, userID, domaintag.MaxTagsPerUser)

	rec := postTag(t, router, cookie, `{"name":"one more","color":"red"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if code, _ := decodeErrorBody(t, rec); code != "TAG_LIMIT_EXCEEDED" {
		t.Errorf("code: got %q, want TAG_LIMIT_EXCEEDED", code)
	}
	if got := tagRowCount(t, userID); got != domaintag.MaxTagsPerUser {
		t.Errorf("tag count: got %d, want %d", got, domaintag.MaxTagsPerUser)
	}
}

func TestRoute_CreateTag_concurrentCreates_doNotExceedLimit(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-race@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	insertRouteTestTags(t, userID, domaintag.MaxTagsPerUser-1)

	const workers = 5
	codes := make([]int, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			codes[i] = postTag(t, router, cookie, fmt.Sprintf(`{"name":"race-%d","color":"red"}`, i)).Code
		})
	}
	wg.Wait()

	created := 0
	for _, c := range codes {
		switch c {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
		default:
			t.Errorf("unexpected status %d", c)
		}
	}
	if created != 1 {
		t.Errorf("expected exactly 1 create to succeed, got %d (codes: %v)", created, codes)
	}
	if got := tagRowCount(t, userID); got != domaintag.MaxTagsPerUser {
		t.Errorf("tag count: got %d, want %d", got, domaintag.MaxTagsPerUser)
	}
}

func getTags(t *testing.T, router *chi.Mux, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/tags", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func insertRouteTestTag(t *testing.T, userID, name, color string, createdAt time.Time) string {
	t.Helper()
	var id string
	if err := routeTestPool.QueryRow(context.Background(),
		"INSERT INTO tags (user_id, name, color, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, name, color, createdAt,
	).Scan(&id); err != nil {
		t.Fatalf("insert tag: %v", err)
	}
	return id
}

func insertRouteTestEventTag(t *testing.T, eventID, tagID string, position int) {
	t.Helper()
	if _, err := routeTestPool.Exec(context.Background(),
		"INSERT INTO event_tags (event_id, tag_id, position) VALUES ($1, $2, $3)",
		eventID, tagID, position,
	); err != nil {
		t.Fatalf("insert event tag: %v", err)
	}
}

func TestRoute_ListTags_unauthenticated_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()

	rec := getTags(t, router, nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_ListTags_noTags_returnsEmptyArray(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-list-empty@ex.com")

	rec := getTags(t, router, cookie)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"tags":[]}` {
		t.Errorf("body: got %s, want {\"tags\":[]}", got)
	}
}

func TestRoute_ListTags_returnsOwnTagsInCreationOrderWithEventCount(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-list@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	otherCookie := signupAndGetCookie(t, router, "tag-list-other@ex.com")
	otherID := getUserIDFromCookie(t, router, otherCookie)

	now := time.Now().UTC()
	// Inserted newest first so that insertion order alone cannot pass.
	newer := insertRouteTestTag(t, userID, "Alpha", "blue", now)
	older := insertRouteTestTag(t, userID, "Beta", "red", now.Add(-time.Hour))
	othersTag := insertRouteTestTag(t, otherID, "Other", "green", now.Add(-2*time.Hour))
	ev1 := insertRouteTestEvent(t, userID, "Meeting", now, now.Add(time.Hour))
	ev2 := insertRouteTestEvent(t, userID, "Lunch", now, now.Add(time.Hour))
	insertRouteTestEventTag(t, ev1, older, 1)
	insertRouteTestEventTag(t, ev2, older, 1)
	insertRouteTestEventTag(t, ev1, newer, 2)
	otherEv := insertRouteTestEvent(t, otherID, "Other", now, now.Add(time.Hour))
	insertRouteTestEventTag(t, otherEv, othersTag, 1)

	rec := getTags(t, router, cookie)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Tags []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			Color      string `json:"color"`
			EventCount int    `json:"eventCount"`
		} `json:"tags"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := []struct {
		id, name, color string
		eventCount      int
	}{
		{older, "Beta", "red", 2},
		{newer, "Alpha", "blue", 1},
	}
	if len(body.Tags) != len(want) {
		t.Fatalf("expected %d tags, got %d: %+v", len(want), len(body.Tags), body.Tags)
	}
	for i, w := range want {
		got := body.Tags[i]
		if got.ID != w.id || got.Name != w.name || got.Color != w.color || got.EventCount != w.eventCount {
			t.Errorf("tags[%d]: got %+v, want %+v", i, got, w)
		}
	}
}

func putTag(t *testing.T, router *chi.Mux, cookie *http.Cookie, id, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/tags/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func tagRow(t *testing.T, id string) (name, color string) {
	t.Helper()
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT name, color FROM tags WHERE id = $1", id,
	).Scan(&name, &color); err != nil {
		t.Fatalf("select tag: %v", err)
	}
	return name, color
}

func TestRoute_UpdateTag_unauthenticated_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()

	rec := putTag(t, router, nil, uuid.NewString(), `{"name":"work","color":"red"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_UpdateTag_success_returns200WithEventCountAndStoresChanges(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-update@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC()
	tagID := insertRouteTestTag(t, userID, "Work", "red", now)
	ev1 := insertRouteTestEvent(t, userID, "Meeting", now, now.Add(time.Hour))
	ev2 := insertRouteTestEvent(t, userID, "Lunch", now, now.Add(time.Hour))
	insertRouteTestEventTag(t, ev1, tagID, 1)
	insertRouteTestEventTag(t, ev2, tagID, 1)

	rec := putTag(t, router, cookie, tagID, `{"name":"  Office  ","color":"indigo"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Color      string `json:"color"`
		EventCount *int   `json:"eventCount"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != tagID || body.Name != "Office" || body.Color != "indigo" {
		t.Errorf("body: got %+v", body)
	}
	if body.EventCount == nil || *body.EventCount != 2 {
		t.Errorf("eventCount: got %v, want 2", body.EventCount)
	}
	if name, color := tagRow(t, tagID); name != "Office" || color != "indigo" {
		t.Errorf("stored tag: got %s/%s, want Office/indigo", name, color)
	}
}

func TestRoute_UpdateTag_ownNameWithDifferentCase_returns200(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-update-case@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	tagID := insertRouteTestTag(t, userID, "work", "red", time.Now().UTC())

	rec := putTag(t, router, cookie, tagID, `{"name":"WORK","color":"red"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if name, _ := tagRow(t, tagID); name != "WORK" {
		t.Errorf("stored name: got %s, want WORK", name)
	}
}

func TestRoute_UpdateTag_nameOfAnotherOwnTag_returns409(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-update-conflict@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)
	now := time.Now().UTC()
	insertRouteTestTag(t, userID, "Work", "red", now)
	tagID := insertRouteTestTag(t, userID, "Private", "blue", now)

	rec := putTag(t, router, cookie, tagID, `{"name":"WORK","color":"blue"}`)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if code, _ := decodeErrorBody(t, rec); code != "TAG_NAME_TAKEN" {
		t.Errorf("code: got %q, want TAG_NAME_TAKEN", code)
	}
	if name, _ := tagRow(t, tagID); name != "Private" {
		t.Errorf("stored name: got %s, want it unchanged as Private", name)
	}
}

func TestRoute_UpdateTag_notOwnedOrMissing_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-update-404@ex.com")
	otherCookie := signupAndGetCookie(t, router, "tag-update-404-other@ex.com")
	otherID := getUserIDFromCookie(t, router, otherCookie)
	othersTag := insertRouteTestTag(t, otherID, "Other", "green", time.Now().UTC())

	for name, id := range map[string]string{"another user's tag": othersTag, "missing tag": uuid.NewString()} {
		t.Run(name, func(t *testing.T) {
			rec := putTag(t, router, cookie, id, `{"name":"Mine","color":"red"}`)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
	if name, color := tagRow(t, othersTag); name != "Other" || color != "green" {
		t.Errorf("another user's tag: got %s/%s, want it unchanged", name, color)
	}
}

func TestRoute_UpdateTag_invalidID_returns400(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-update-400@ex.com")

	rec := putTag(t, router, cookie, "not-a-uuid", `{"name":"work","color":"red"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func deleteTag(t *testing.T, router *chi.Mux, cookie *http.Cookie, id string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/tags/"+id, http.NoBody)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func rowExists(t *testing.T, query string, args ...any) bool {
	t.Helper()
	var exists bool
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT EXISTS ("+query+")", args...,
	).Scan(&exists); err != nil {
		t.Fatalf("check row existence: %v", err)
	}
	return exists
}

func TestRoute_DeleteTag_unauthenticated_returns401(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()

	rec := deleteTag(t, router, nil, uuid.NewString())

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRoute_DeleteTag_success_returns204AndDetachesTagFromEvents(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-delete@ex.com")
	userID := getUserIDFromCookie(t, router, cookie)

	now := time.Now().UTC()
	tagID := insertRouteTestTag(t, userID, "Work", "red", now)
	keptTagID := insertRouteTestTag(t, userID, "Private", "blue", now)
	eventID := insertRouteTestEvent(t, userID, "Meeting", now, now.Add(time.Hour))
	insertRouteTestEventTag(t, eventID, tagID, 1)
	insertRouteTestEventTag(t, eventID, keptTagID, 2)

	rec := deleteTag(t, router, cookie, tagID)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Errorf("expected an empty body, got %q", rec.Body.String())
	}
	if rowExists(t, "SELECT 1 FROM tags WHERE id = $1", tagID) {
		t.Error("expected the tag to be deleted")
	}
	if rowExists(t, "SELECT 1 FROM event_tags WHERE tag_id = $1", tagID) {
		t.Error("expected the tag to be detached from the event")
	}
	if !rowExists(t, "SELECT 1 FROM events WHERE id = $1", eventID) {
		t.Error("expected the event to be kept")
	}
	if !rowExists(t, "SELECT 1 FROM event_tags WHERE event_id = $1 AND tag_id = $2", eventID, keptTagID) {
		t.Error("expected the event's other tag to stay attached")
	}
}

func TestRoute_DeleteTag_notOwnedOrMissing_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-delete-404@ex.com")
	otherCookie := signupAndGetCookie(t, router, "tag-delete-404-other@ex.com")
	otherID := getUserIDFromCookie(t, router, otherCookie)
	othersTag := insertRouteTestTag(t, otherID, "Other", "green", time.Now().UTC())

	for name, id := range map[string]string{"another user's tag": othersTag, "missing tag": uuid.NewString()} {
		t.Run(name, func(t *testing.T) {
			rec := deleteTag(t, router, cookie, id)

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
	if !rowExists(t, "SELECT 1 FROM tags WHERE id = $1", othersTag) {
		t.Error("expected another user's tag to be kept")
	}
}

func TestRoute_DeleteTag_invalidID_returns400(t *testing.T) {
	setupRouteTest(t)
	router := buildTagTestRouter()
	cookie := signupAndGetCookie(t, router, "tag-delete-400@ex.com")

	rec := deleteTag(t, router, cookie, "not-a-uuid")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
