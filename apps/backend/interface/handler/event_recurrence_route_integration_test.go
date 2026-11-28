//go:build integration

package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type routeTestEventWithSeries struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	StartAt         time.Time       `json:"startAt"`
	RecurringSeries json.RawMessage `json:"recurringSeries"`
}

// postDailyCount3Series creates a daily/count=3 series starting at start and
// returns the created first occurrence.
func postDailyCount3Series(
	t *testing.T, router *chi.Mux, cookie *http.Cookie, start time.Time,
) routeTestEventWithSeries {
	t.Helper()
	body := fmt.Sprintf(`{
		"title": "Daily series",
		"startAt": "%s",
		"endAt":   "%s",
		"recurrence": {
			"pattern":  {"type": "daily"},
			"end":      {"type": "count", "count": 3},
			"timezone": "Asia/Tokyo"
		}
	}`, start.Format(time.RFC3339), start.Add(time.Hour).Format(time.RFC3339))

	req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /events expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created routeTestEventWithSeries
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	return created
}

func getEventsWithSeries(
	t *testing.T, router *chi.Mux, cookie *http.Cookie, start, end time.Time,
) []routeTestEventWithSeries {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, eventsRangePath(start, end), nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /events expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Events []routeTestEventWithSeries `json:"events"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode GET response: %v", err)
	}
	return body.Events
}

// GET /events returns, for every occurrence, the same recurringSeries that
// POST /events returned when the series was created.
func TestRoute_GetEvents_returnsRecurringSeriesOfEachOccurrence(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "get-events-series@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	created := postDailyCount3Series(t, router, cookie, start)
	oneOffStart := start.Add(3 * time.Hour)
	oneOffID := insertRouteTestEvent(t, getUserIDFromCookie(t, router, cookie), "One-off",
		oneOffStart, oneOffStart.Add(time.Hour))

	events := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 3))
	if len(events) != 4 {
		t.Fatalf("expected 3 occurrences and 1 one-off event, got %d", len(events))
	}
	for _, e := range events {
		if e.ID == oneOffID {
			if string(e.RecurringSeries) != jsonNull {
				t.Errorf("expected null recurringSeries for one-off event, got %s", e.RecurringSeries)
			}
			continue
		}
		assertJSONEqual(t, e.RecurringSeries, string(created.RecurringSeries))
	}
}

// A subscriber must not see the owner's series of a subscribed occurrence.
func TestRoute_GetEvents_subscribedOccurrence_returnsNullRecurringSeries(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	ownerCookie := signupAndGetCookie(t, router, "series-owner@ex.com")
	subscriberCookie := signupAndGetCookie(t, router, "series-subscriber@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	created := postDailyCount3Series(t, router, ownerCookie, start)
	insertRouteTestEventSubscription(t, created.ID, getUserIDFromCookie(t, router, subscriberCookie))

	events := getEventsWithSeries(t, router, subscriberCookie, start.Add(-time.Hour), start.AddDate(0, 0, 3))
	if len(events) != 1 {
		t.Fatalf("expected only the subscribed occurrence, got %d", len(events))
	}
	if events[0].ID != created.ID {
		t.Fatalf("expected subscribed occurrence %s, got %s", created.ID, events[0].ID)
	}
	if string(events[0].RecurringSeries) != jsonNull {
		t.Errorf("expected null recurringSeries for subscribed event, got %s", events[0].RecurringSeries)
	}
}

// putEventWithSeries sends PUT /events/{id} with times around start plus the
// extra JSON members, and returns the recorder.
func putEventWithSeries(
	t *testing.T, router *chi.Mux, cookie *http.Cookie, id string, start time.Time, extra string,
) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"title": "Edited", "startAt": "%s", "endAt": "%s"%s}`,
		start.Format(time.RFC3339), start.Add(time.Hour).Format(time.RFC3339), extra)
	req := httptest.NewRequest(http.MethodPut, "/events/"+id, strings.NewReader(body))
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeEventWithSeries(t *testing.T, rec *httptest.ResponseRecorder) routeTestEventWithSeries {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("PUT /events/{id} expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var ev routeTestEventWithSeries
	if err := json.NewDecoder(rec.Body).Decode(&ev); err != nil {
		t.Fatalf("decode PUT response: %v", err)
	}
	return ev
}

func countRouteTestSeries(t *testing.T) int {
	t.Helper()
	var n int
	if err := routeTestPool.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM recurring_event_series").Scan(&n); err != nil {
		t.Fatalf("count series: %v", err)
	}
	return n
}

// A recurrence on a one-off event makes it the first occurrence of a series.
func TestRoute_UpdateEvent_oneOffWithRecurrence_convertsToSeries(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-convert@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	id := insertRouteTestEvent(t, getUserIDFromCookie(t, router, cookie), "One-off", start, start.Add(time.Hour))

	updated := decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, id, start,
		`, "scope": "single", "recurrence": {"pattern": {"type": "daily"}, "end": {"type": "count", "count": 3}, "timezone": "Asia/Tokyo"}`))
	if updated.ID != id {
		t.Errorf("expected the event to keep its id %s, got %s", id, updated.ID)
	}
	if string(updated.RecurringSeries) == jsonNull {
		t.Fatal("expected recurringSeries in the response, got null")
	}

	events := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 3))
	if len(events) != 3 {
		t.Fatalf("expected 3 occurrences, got %d", len(events))
	}
	if events[0].ID != id {
		t.Errorf("expected the first occurrence to be %s, got %s", id, events[0].ID)
	}
	for _, e := range events {
		if e.Title != "Edited" {
			t.Errorf("occurrence %s: title %q", e.ID, e.Title)
		}
		assertJSONEqual(t, e.RecurringSeries, string(updated.RecurringSeries))
	}
}

// "single" detaches only the target, and the other occurrences stay.
func TestRoute_UpdateEvent_singleScope_detachesOccurrence(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-single@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	created := postDailyCount3Series(t, router, cookie, start)
	before := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 3))
	second := before[1]

	// Moving it to another day is allowed for "single".
	moved := second.StartAt.AddDate(0, 0, 5)
	updated := decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, second.ID, moved, `, "scope": "single"`))
	if string(updated.RecurringSeries) != jsonNull {
		t.Errorf("expected null recurringSeries for the detached event, got %s", updated.RecurringSeries)
	}

	after := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 10))
	if len(after) != 3 {
		t.Fatalf("expected 2 occurrences and the detached event, got %d", len(after))
	}
	for _, e := range after {
		if e.ID == second.ID {
			if string(e.RecurringSeries) != jsonNull || !e.StartAt.Equal(moved) || e.Title != "Edited" {
				t.Errorf("detached event: %+v", e)
			}
			continue
		}
		if e.Title != "Daily series" {
			t.Errorf("occurrence %s changed: title %q", e.ID, e.Title)
		}
		assertJSONEqual(t, e.RecurringSeries, string(created.RecurringSeries))
	}
}

// Detaching every occurrence deletes the series.
func TestRoute_UpdateEvent_singleScope_lastOccurrence_deletesSeries(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-single-last@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	postDailyCount3Series(t, router, cookie, start)
	for _, e := range getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 3)) {
		decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, e.ID, e.StartAt, `, "scope": "single"`))
	}

	if n := countRouteTestSeries(t); n != 0 {
		t.Errorf("expected the empty series to be deleted, got %d series", n)
	}
}

// An occurrence needs a scope.
func TestRoute_UpdateEvent_occurrenceWithoutScope_returns400(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-no-scope@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	created := postDailyCount3Series(t, router, cookie, start)

	rec := putEventWithSeries(t, router, cookie, created.ID, start, "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"code":"REQUIRED"`) || !strings.Contains(rec.Body.String(), `"field":"scope"`) {
		t.Errorf("expected scope REQUIRED in details, got %s", rec.Body.String())
	}
}

// Another user's occurrence is reported as missing.
func TestRoute_UpdateEvent_otherUsersOccurrence_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	ownerCookie := signupAndGetCookie(t, router, "put-series-owner@ex.com")
	otherCookie := signupAndGetCookie(t, router, "put-series-other@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	created := postDailyCount3Series(t, router, ownerCookie, start)

	rec := putEventWithSeries(t, router, otherCookie, created.ID, start, `, "scope": "single"`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if n := countRouteTestSeries(t); n != 1 {
		t.Errorf("expected the owner's series to remain, got %d series", n)
	}
}
