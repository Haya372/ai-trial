//go:build integration

package handler_test

import (
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
