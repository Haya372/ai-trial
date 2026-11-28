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

type routeTestSeries struct {
	ID  string `json:"id"`
	End struct {
		Type  string `json:"type"`
		Count *int   `json:"count"`
	} `json:"end"`
}

func decodeSeries(t *testing.T, raw json.RawMessage) routeTestSeries {
	t.Helper()
	var s routeTestSeries
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("decode recurringSeries %s: %v", raw, err)
	}
	return s
}

// postDailyCount5Series creates a daily/count=5 series and returns its
// occurrences in order of start.
func postDailyCount5Series(
	t *testing.T, router *chi.Mux, cookie *http.Cookie, start time.Time,
) []routeTestEventWithSeries {
	t.Helper()
	body := fmt.Sprintf(`{
		"title": "Daily series",
		"startAt": "%s",
		"endAt":   "%s",
		"recurrence": {
			"pattern":  {"type": "daily"},
			"end":      {"type": "count", "count": 5},
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
	occs := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 5))
	if len(occs) != 5 {
		t.Fatalf("expected 5 occurrences, got %d", len(occs))
	}
	return occs
}

// "future" without a recurrence keeps the rule: the occurrences from the
// target on move to a new series with the remaining count and keep their ids.
func TestRoute_UpdateEvent_futureScope_keepingRule_splitsSeries(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-future@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	before := postDailyCount5Series(t, router, cookie, start)
	oldSeries := decodeSeries(t, before[0].RecurringSeries)

	// Two hours later on the same date in Asia/Tokyo.
	newStart := before[2].StartAt.Add(2 * time.Hour)
	updated := decodeEventWithSeries(t,
		putEventWithSeries(t, router, cookie, before[2].ID, newStart, `, "scope": "future"`))
	newSeries := decodeSeries(t, updated.RecurringSeries)
	if updated.ID != before[2].ID || newSeries.ID == oldSeries.ID {
		t.Fatalf("expected the target %s in a new series, got %s in %s", before[2].ID, updated.ID, newSeries.ID)
	}
	if newSeries.End.Type != "count" || newSeries.End.Count == nil || *newSeries.End.Count != 3 {
		t.Errorf("expected the new series to have count 3, got %+v", newSeries.End)
	}

	after := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 5))
	if len(after) != 5 {
		t.Fatalf("expected 5 events, got %d", len(after))
	}
	for i, e := range after {
		if e.ID != before[i].ID {
			t.Errorf("event %d: id %s, want %s", i, e.ID, before[i].ID)
		}
		s := decodeSeries(t, e.RecurringSeries)
		if i < 2 {
			if e.Title != "Daily series" || !e.StartAt.Equal(before[i].StartAt) || s.ID != oldSeries.ID {
				t.Errorf("event %d before the target changed: %+v", i, e)
			}
			continue
		}
		if e.Title != "Edited" || !e.StartAt.Equal(before[i].StartAt.Add(2*time.Hour)) || s.ID != newSeries.ID {
			t.Errorf("event %d from the target: %+v", i, e)
		}
	}
}

// "all" updates every occurrence in place but leaves a detached event alone.
func TestRoute_UpdateEvent_allScope_keepingRule_updatesEveryOccurrence(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-all@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	before := postDailyCount5Series(t, router, cookie, start)
	decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, before[1].ID, before[1].StartAt.AddDate(0, 0, 10),
		`, "scope": "single"`))

	newStart := before[3].StartAt.Add(-30 * time.Minute)
	updated := decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, before[3].ID, newStart, `, "scope": "all"`))
	if updated.ID != before[3].ID || !updated.StartAt.Equal(newStart) {
		t.Errorf("expected the target at %v, got %+v", newStart, updated)
	}

	after := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 12))
	if len(after) != 5 {
		t.Fatalf("expected 4 occurrences and the detached event, got %d", len(after))
	}
	for _, e := range after {
		if e.ID == before[1].ID {
			if e.Title != "Edited" || string(e.RecurringSeries) != jsonNull ||
				!e.StartAt.Equal(before[1].StartAt.AddDate(0, 0, 10)) {
				t.Errorf("detached event changed: %+v", e)
			}
			continue
		}
		var orig routeTestEventWithSeries
		for _, b := range before {
			if b.ID == e.ID {
				orig = b
			}
		}
		if orig.ID == "" || e.Title != "Edited" || !e.StartAt.Equal(orig.StartAt.Add(-30*time.Minute)) {
			t.Errorf("occurrence %s: %+v, want the original moved 30 minutes earlier", e.ID, e)
		}
	}
}

// "all" with a new rule regenerates the series from its first occurrence.
func TestRoute_UpdateEvent_allScope_newRule_regeneratesOccurrences(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-all-rule@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	before := postDailyCount5Series(t, router, cookie, start)
	oldSeries := decodeSeries(t, before[0].RecurringSeries)

	updated := decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, before[2].ID, before[2].StartAt,
		`, "scope": "all", "recurrence": {"pattern": {"type": "weekly"},`+
			` "end": {"type": "count", "count": 2}, "timezone": "Asia/Tokyo"}`))
	if !updated.StartAt.Equal(start) || decodeSeries(t, updated.RecurringSeries).ID != oldSeries.ID {
		t.Errorf("expected the regenerated first occurrence at %v in the same series, got %+v", start, updated)
	}

	after := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 14))
	if len(after) != 2 || !after[0].StartAt.Equal(start) || !after[1].StartAt.Equal(start.AddDate(0, 0, 7)) {
		t.Fatalf("expected 2 weekly occurrences from %v, got %+v", start, after)
	}
	for _, e := range after {
		for _, b := range before {
			if e.ID == b.ID {
				t.Errorf("expected new ids, got the old %s", e.ID)
			}
		}
	}
}

// Keeping the rule keeps the date: moving it would break the weekday and the
// excluded dates of the series.
func TestRoute_UpdateEvent_futureOrAllScope_dateChanged_returns400(t *testing.T) {
	for _, scope := range []string{"future", "all"} {
		t.Run(scope, func(t *testing.T) {
			setupRouteTest(t)
			router := buildEventTestRouter()
			cookie := signupAndGetCookie(t, router, "put-date-"+scope+"@ex.com")

			start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
			before := postDailyCount5Series(t, router, cookie, start)

			rec := putEventWithSeries(t, router, cookie, before[2].ID, before[2].StartAt.Add(24*time.Hour),
				`, "scope": "`+scope+`"`)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"code":"DATE_CHANGE_NOT_ALLOWED"`) ||
				!strings.Contains(rec.Body.String(), `"field":"startAt"`) {
				t.Errorf("expected startAt DATE_CHANGE_NOT_ALLOWED in details, got %s", rec.Body.String())
			}
		})
	}
}

// "future" with null turns the target into a one-off event and ends the
// series before it.
func TestRoute_UpdateEvent_futureScope_nullRecurrence_endsSeries(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	cookie := signupAndGetCookie(t, router, "put-future-null@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	before := postDailyCount5Series(t, router, cookie, start)

	updated := decodeEventWithSeries(t, putEventWithSeries(t, router, cookie, before[2].ID, before[2].StartAt,
		`, "scope": "future", "recurrence": null`))
	if string(updated.RecurringSeries) != jsonNull {
		t.Errorf("expected a one-off event, got %s", updated.RecurringSeries)
	}

	after := getEventsWithSeries(t, router, cookie, start.Add(-time.Hour), start.AddDate(0, 0, 5))
	if len(after) != 3 {
		t.Fatalf("expected 2 occurrences and the one-off event, got %d", len(after))
	}
	for i, e := range after {
		if e.ID != before[i].ID {
			t.Errorf("event %d: id %s, want %s", i, e.ID, before[i].ID)
		}
	}
	if string(after[2].RecurringSeries) != jsonNull || string(after[0].RecurringSeries) == jsonNull {
		t.Errorf("expected the earlier occurrences to stay in the series: %+v", after)
	}
}

func TestRoute_UpdateEvent_allScope_otherUsersOccurrence_returns404(t *testing.T) {
	setupRouteTest(t)
	router := buildEventTestRouter()
	ownerCookie := signupAndGetCookie(t, router, "put-all-owner@ex.com")
	otherCookie := signupAndGetCookie(t, router, "put-all-other@ex.com")

	start := time.Date(2027, 1, 4, 1, 0, 0, 0, time.UTC)
	before := postDailyCount5Series(t, router, ownerCookie, start)

	rec := putEventWithSeries(t, router, otherCookie, before[0].ID, start, `, "scope": "all"`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	for _, e := range getEventsWithSeries(t, router, ownerCookie, start.Add(-time.Hour), start.AddDate(0, 0, 5)) {
		if e.Title != "Daily series" {
			t.Errorf("owner's occurrence %s changed: %q", e.ID, e.Title)
		}
	}
}
