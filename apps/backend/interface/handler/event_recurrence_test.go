package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain"
	domainevent "github.com/Haya372/ai-trial/backend/domain/event"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
	"github.com/Haya372/ai-trial/backend/interface/ctxkey"
	"github.com/Haya372/ai-trial/backend/interface/handler"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

// mustBuildWeeklyCount2Series returns a simple RecurringEventSeries for
// weekly/count=2 starting at startAt.
func mustBuildWeeklyCount2Series(t *testing.T, userID uuid.UUID) rs.RecurringEventSeries {
	t.Helper()
	pattern, err := rs.NewRecurrencePattern(valWeekly, nil)
	if err != nil {
		t.Fatalf("NewRecurrencePattern: %v", err)
	}
	end, err := rs.NewCountEnd(2)
	if err != nil {
		t.Fatalf("NewCountEnd: %v", err)
	}
	tz, err := rs.NewSeriesTimezone(valUTC)
	if err != nil {
		t.Fatalf("NewSeriesTimezone: %v", err)
	}
	series, err := rs.New(uuid.New(), userID, rs.RecurrenceRule{Pattern: pattern, End: end, Timezone: tz}, nil)
	if err != nil {
		t.Fatalf("rs.New: %v", err)
	}
	return series
}

// TestEventHandler_CreateEvent_RecurrenceInput_MappedCorrectly verifies that the
// handler converts the JSON body's recurrence object into RecurrenceInput.
func TestEventHandler_CreateEvent_RecurrenceInput_MappedCorrectly(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	var captured eventuc.CreateEventInput
	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			captured = in
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence: map[string]any{
			keyPattern: map[string]any{
				keyType:      valMonthlyNthWeekday,
				"monthlyNth": "2",
			},
			keyEnd: map[string]any{
				keyType:     "until_date",
				"untilDate": "2026-12-31",
			},
			keyTimezone: valAsiaTokyoTZ,
		},
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if captured.Recurrence == nil {
		t.Fatal("expected Recurrence in input, got nil")
	}
	if captured.Recurrence.PatternType != "monthly_nth_weekday" {
		t.Errorf("PatternType mismatch: got %q", captured.Recurrence.PatternType)
	}
	if captured.Recurrence.MonthlyNth == nil || *captured.Recurrence.MonthlyNth != "2" {
		t.Errorf("MonthlyNth mismatch: got %v", captured.Recurrence.MonthlyNth)
	}
	if captured.Recurrence.EndType != valUntilDate {
		t.Errorf("EndType mismatch: got %q", captured.Recurrence.EndType)
	}
	if captured.Recurrence.UntilDate == nil || *captured.Recurrence.UntilDate != "2026-12-31" {
		t.Errorf("UntilDate mismatch: got %v", captured.Recurrence.UntilDate)
	}
	if captured.Recurrence.Timezone != valAsiaTokyoTZ {
		t.Errorf("Timezone mismatch: got %q", captured.Recurrence.Timezone)
	}
}

// TestEventHandler_CreateEvent_RecurrenceCount_MappedCorrectly verifies that
// count is passed through.
func TestEventHandler_CreateEvent_RecurrenceCount_MappedCorrectly(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	var captured eventuc.CreateEventInput
	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			captured = in
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence: map[string]any{
			keyPattern:  map[string]any{keyType: valWeekly},
			keyEnd:      map[string]any{keyType: valCount, valCount: 5},
			keyTimezone: valUTC,
		},
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if captured.Recurrence == nil {
		t.Fatal("expected Recurrence, got nil")
	}
	if captured.Recurrence.Count == nil || *captured.Recurrence.Count != 5 {
		t.Errorf("Count mismatch: got %v", captured.Recurrence.Count)
	}
}

// TestEventHandler_CreateEvent_RecurrenceOmitted_NilInput verifies that omitting
// recurrence passes nil to the usecase.
func TestEventHandler_CreateEvent_RecurrenceOmitted_NilInput(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	var captured eventuc.CreateEventInput
	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			captured = in
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if captured.Recurrence != nil {
		t.Errorf("expected nil Recurrence for omitted field, got %v", captured.Recurrence)
	}
}

// TestEventHandler_CreateEvent_RecurrenceNull_NilInput verifies that null
// recurrence passes nil to the usecase.
func TestEventHandler_CreateEvent_RecurrenceNull_NilInput(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	var captured eventuc.CreateEventInput
	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			captured = in
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence:  nil,
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if captured.Recurrence != nil {
		t.Errorf("expected nil Recurrence for null field, got %v", captured.Recurrence)
	}
}

// TestEventHandler_CreateEvent_SeriesPresent_ReturnsRecurringSeries verifies that
// when the usecase returns a Series, the response contains recurringSeries with the
// correct id/pattern/end/timezone.
func TestEventHandler_CreateEvent_SeriesPresent_ReturnsRecurringSeries(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	series := mustBuildWeeklyCount2Series(t, userID)

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}, Series: series}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence: map[string]any{
			keyPattern:  map[string]any{keyType: valWeekly},
			keyEnd:      map[string]any{keyType: valCount, valCount: 2},
			keyTimezone: valUTC,
		},
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		RecurringSeries *struct {
			ID      string `json:"id"`
			Pattern struct {
				Type string `json:"type"`
			} `json:"pattern"`
			End struct {
				Type  string `json:"type"`
				Count *int   `json:"count"`
			} `json:"end"`
			Timezone string `json:"timezone"`
		} `json:"recurringSeries"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.RecurringSeries == nil {
		t.Fatal("expected recurringSeries in response, got null")
	}
	if resp.RecurringSeries.ID != series.ID().String() {
		t.Errorf("id mismatch: got %q, want %q", resp.RecurringSeries.ID, series.ID())
	}
	if resp.RecurringSeries.Pattern.Type != valWeekly {
		t.Errorf("pattern.type mismatch: got %q", resp.RecurringSeries.Pattern.Type)
	}
	if resp.RecurringSeries.End.Type != valCount {
		t.Errorf("end.type mismatch: got %q", resp.RecurringSeries.End.Type)
	}
	if resp.RecurringSeries.End.Count == nil || *resp.RecurringSeries.End.Count != 2 {
		t.Errorf("end.count mismatch: got %v", resp.RecurringSeries.End.Count)
	}
	if resp.RecurringSeries.Timezone != valUTC {
		t.Errorf("timezone mismatch: got %q", resp.RecurringSeries.Timezone)
	}
}

// TestEventHandler_CreateEvent_SeriesNil_RecurringSeriesNullInResponse verifies that
// when Series is nil, the response JSON contains "recurringSeries": null (key present).
func TestEventHandler_CreateEvent_SeriesNil_RecurringSeriesNullInResponse(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}, Series: nil}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	// The key "recurringSeries" must be present with value null
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	val, ok := raw["recurringSeries"]
	if !ok {
		t.Fatal("key recurringSeries missing from response")
	}
	if string(val) != jsonNull {
		t.Errorf("expected null, got %s", val)
	}
}

// TestEventHandler_UpdateEvent_RecurringSeriesNullInResponse verifies that
// PUT /events/{id} always returns "recurringSeries": null (#404 scope).
func TestEventHandler_UpdateEvent_RecurringSeriesNullInResponse(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now().UTC()

	updatedEv, _ := domainevent.New(eventID, userID, "Title", "", now, now.Add(time.Hour), "", "")
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{Event: updatedEv, Tags: []eventuc.EventTagView{}}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)

	body := `{"title":"Title","startAt":"` + now.Format(time.RFC3339) +
		`","endAt":"` + now.Add(time.Hour).Format(time.RFC3339) + `"}`
	req := putEventRequestAsUser(t, eventID.String(), body)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&raw); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	val, ok := raw["recurringSeries"]
	if !ok {
		t.Fatal("key recurringSeries missing from PUT response")
	}
	if string(val) != jsonNull {
		t.Errorf("expected null for PUT response, got %s", val)
	}
}

// TestEventHandler_GetEvents_RecurringSeriesNullInResponse verifies that
// GET /events returns "recurringSeries": null (key present) for events without
// a series, i.e. one-off and subscribed events.
func TestEventHandler_GetEvents_RecurringSeriesNullInResponse(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()

	stub := &stubListEventsExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
			return []eventuc.EventReadModel{
				{ID: uuid.New(), Title: "Event", StartAt: now, EndAt: now.Add(time.Hour)},
			}, nil
		},
	}
	h := handler.NewEventHandler(stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.GetEvents(w, req, testGetEventsParams(t))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var body struct {
		Events []map[string]json.RawMessage `json:"events"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(body.Events))
	}
	val, ok := body.Events[0]["recurringSeries"]
	if !ok {
		t.Fatal("key recurringSeries missing from GET events response")
	}
	if string(val) != jsonNull {
		t.Errorf("expected null for GET events, got %s", val)
	}
}

func TestEventHandler_GetEvents_ReturnsRecurringSeries(t *testing.T) {
	seriesID := uuid.New()
	nth, untilDate, count := "last", "2027-03-31", 1

	tests := []struct {
		name   string
		series eventuc.RecurringSeriesReadModel
		want   string
	}{
		{
			name: "count end",
			series: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: valWeekly, EndType: valCount, Count: &count, Timezone: valAsiaTokyoTZ,
			},
			want: `{"id":"` + seriesID.String() + `","pattern":{"type":"weekly"},` +
				`"end":{"type":"count","count":1},"timezone":"Asia/Tokyo"}`,
		},
		{
			name: "monthly nth weekday until date",
			series: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: valMonthlyNthWeekday, MonthlyNth: &nth,
				EndType: valUntilDate, UntilDate: &untilDate, Timezone: valUTC,
			},
			want: `{"id":"` + seriesID.String() + `","pattern":{"type":"monthly_nth_weekday","monthlyNth":"last"},` +
				`"end":{"type":"until_date","untilDate":"2027-03-31"},"timezone":"UTC"}`,
		},
		{
			name: "indefinite end",
			series: eventuc.RecurringSeriesReadModel{
				ID: seriesID, PatternType: "daily", EndType: "indefinite", Timezone: valUTC,
			},
			want: `{"id":"` + seriesID.String() + `","pattern":{"type":"daily"},` +
				`"end":{"type":"indefinite"},"timezone":"UTC"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			stub := &stubListEventsExec{
				fn: func(_ context.Context, _ uuid.UUID, _ eventuc.ListEventsInput) ([]eventuc.EventReadModel, error) {
					series := tt.series
					return []eventuc.EventReadModel{
						{ID: uuid.New(), Title: "Event", StartAt: now, EndAt: now.Add(time.Hour), RecurringSeries: &series},
					}, nil
				},
			}
			h := handler.NewEventHandler(
				stub, &stubCreateEventExec{}, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger,
			)

			req := getEventsRequest(t, testRangeStartAt, testRangeEndAt)
			user := newStubUser(uuid.New(), "u@ex.com", "U")
			req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, user))
			w := httptest.NewRecorder()

			h.GetEvents(w, req, testGetEventsParams(t))

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			var body struct {
				Events []struct {
					RecurringSeries json.RawMessage `json:"recurringSeries"`
				} `json:"events"`
			}
			if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if len(body.Events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(body.Events))
			}
			assertJSONEqual(t, body.Events[0].RecurringSeries, tt.want)
		})
	}
}

func assertJSONEqual(t *testing.T, got json.RawMessage, want string) {
	t.Helper()
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("unmarshal got %s: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("unmarshal want: %v", err)
	}
	if !reflect.DeepEqual(g, w) {
		t.Errorf("JSON mismatch:\n got: %s\nwant: %s", got, want)
	}
}

// TestEventHandler_CreateEvent_RecurrenceValidationError_Returns400WithDetails verifies
// that ValidationError from recurrence validation is correctly returned as 400.
func TestEventHandler_CreateEvent_RecurrenceValidationError_Returns400WithDetails(t *testing.T) {
	userID := uuid.New()

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, _ eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			return eventuc.CreateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "recurrence.end.count", Code: "OUT_OF_RANGE", Message: "count must be at least 2"},
			}}
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	now := time.Now().UTC()
	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence: map[string]any{
			keyPattern:  map[string]any{keyType: valWeekly},
			keyEnd:      map[string]any{keyType: valCount, valCount: 1},
			keyTimezone: valUTC,
		},
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var respBody map[string]any
	_ = json.NewDecoder(w.Body).Decode(&respBody)
	if respBody["code"] != codeValidationError {
		t.Errorf("expected VALIDATION_ERROR, got %v", respBody["code"])
	}
	details, ok := respBody["details"].([]any)
	if !ok || len(details) == 0 {
		t.Fatalf("expected details in response: %v", respBody)
	}
	detail, _ := details[0].(map[string]any)
	if detail["field"] != "recurrence.end.count" {
		t.Errorf("field mismatch: got %v", detail["field"])
	}
	if detail["code"] != "OUT_OF_RANGE" {
		t.Errorf("code mismatch: got %v", detail["code"])
	}
}

func TestEventHandler_CreateEvent_MonthlyNthWeekdaySeries_ReturnsMonthlyNth(t *testing.T) {
	userID := uuid.New()
	nth := "last"
	pattern, err := rs.NewRecurrencePattern(valMonthlyNthWeekday, &nth)
	if err != nil {
		t.Fatalf("NewRecurrencePattern: %v", err)
	}
	tz, err := rs.NewSeriesTimezone(valUTC)
	if err != nil {
		t.Fatalf("NewSeriesTimezone: %v", err)
	}
	series, err := rs.New(uuid.New(), userID,
		rs.RecurrenceRule{Pattern: pattern, End: rs.NewIndefiniteEnd(), Timezone: tz}, nil)
	if err != nil {
		t.Fatalf("rs.New: %v", err)
	}

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}, Series: series}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	now := time.Date(2026, 11, 30, 10, 0, 0, 0, time.UTC)
	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		RecurringSeries json.RawMessage `json:"recurringSeries"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertJSONEqual(t, resp.RecurringSeries, `{"id":"`+series.ID().String()+`",`+
		`"pattern":{"type":"monthly_nth_weekday","monthlyNth":"last"},`+
		`"end":{"type":"indefinite"},"timezone":"UTC"}`)
}

// TestEventHandler_CreateEvent_UntilDateSeries_ReturnsRecurringSeries tests that
// an until_date series is correctly serialized in the response (untilDate format).
func TestEventHandler_CreateEvent_UntilDateSeries_ReturnsRecurringSeries(t *testing.T) {
	userID := uuid.New()
	now := time.Date(2026, 11, 23, 10, 0, 0, 0, time.UTC)

	// Build a series with until_date end
	pattern, _ := rs.NewRecurrencePattern("daily", nil)
	ld, _ := rs.NewLocalDate(2027, 1, 31)
	end := rs.NewUntilDateEnd(ld)
	tz, _ := rs.NewSeriesTimezone(valAsiaTokyoTZ)
	series, _ := rs.New(uuid.New(), userID, rs.RecurrenceRule{Pattern: pattern, End: end, Timezone: tz}, nil)

	stub := &stubCreateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.CreateEventInput) (eventuc.CreateEventOutput, error) {
			ev, _ := domainevent.New(uuid.New(), userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.CreateEventOutput{Event: ev, Tags: []eventuc.EventTagView{}, Series: series}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, stub, &stubUpdateEventExec{}, &stubDeleteEventExec{}, testLogger)

	body := map[string]any{
		testKeyTitle:   testEventTitle,
		testKeyStartAt: now.Format(time.RFC3339),
		testKeyEndAt:   now.Add(time.Hour).Format(time.RFC3339),
		keyRecurrence: map[string]any{
			keyPattern:  map[string]any{keyType: "daily"},
			keyEnd:      map[string]any{keyType: valUntilDate, "untilDate": "2027-01-31"},
			keyTimezone: valAsiaTokyoTZ,
		},
	}
	req := createEventRequest(t, body)
	req = req.WithContext(context.WithValue(req.Context(), ctxkey.User, newStubUser(userID, "u@ex.com", "U")))
	w := httptest.NewRecorder()

	h.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		RecurringSeries *struct {
			End struct {
				Type      string  `json:"type"`
				UntilDate *string `json:"untilDate"`
				Count     *int    `json:"count"`
			} `json:"end"`
			Timezone string `json:"timezone"`
		} `json:"recurringSeries"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.RecurringSeries == nil {
		t.Fatal("expected recurringSeries, got null")
	}
	if resp.RecurringSeries.End.Type != valUntilDate {
		t.Errorf("end.type mismatch: got %q", resp.RecurringSeries.End.Type)
	}
	if resp.RecurringSeries.End.UntilDate == nil || *resp.RecurringSeries.End.UntilDate != "2027-01-31" {
		t.Errorf("end.untilDate mismatch: got %v", resp.RecurringSeries.End.UntilDate)
	}
	if resp.RecurringSeries.End.Count != nil {
		t.Errorf("end.count should be absent for until_date, got %v", *resp.RecurringSeries.End.Count)
	}
	if resp.RecurringSeries.Timezone != valAsiaTokyoTZ {
		t.Errorf("timezone mismatch: got %q", resp.RecurringSeries.Timezone)
	}
}

// TestEventHandler_UpdateEvent_RecurrenceAndScope_Mapped verifies that the
// handler tells omitted, null and object recurrence apart and passes scope.
func TestEventHandler_UpdateEvent_RecurrenceAndScope_Mapped(t *testing.T) {
	const times = `"title":"x","startAt":"2026-09-01T00:00:00Z","endAt":"2026-09-01T01:00:00Z"`
	tests := []struct {
		name      string
		body      string
		wantInput eventuc.UpdateEventInput
	}{
		{"omitted", `{` + times + `}`, eventuc.UpdateEventInput{}},
		{
			"null with single",
			`{` + times + `,"recurrence":null,"scope":"single"}`,
			eventuc.UpdateEventInput{Recurrence: eventuc.RecurrenceUpdate{Present: true}, Scope: "single"},
		},
		{
			"object",
			`{` + times + `,"recurrence":{"pattern":{"type":"monthly_nth_weekday","monthlyNth":"last"},` +
				`"end":{"type":"count","count":3},"timezone":"Asia/Tokyo"},"scope":"future"}`,
			eventuc.UpdateEventInput{
				Recurrence: eventuc.RecurrenceUpdate{Present: true, Spec: &eventuc.RecurrenceInput{
					PatternType: "monthly_nth_weekday", MonthlyNth: new("last"),
					EndType: "count", Count: new(3), Timezone: "Asia/Tokyo",
				}},
				Scope: "future",
			},
		},
		// The command, not the handler, rejects a value outside the enum.
		{"unknown scope", `{` + times + `,"scope":"weekly"}`, eventuc.UpdateEventInput{Scope: "weekly"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var captured eventuc.UpdateEventInput
			stub := &stubUpdateEventExec{
				fn: func(_ context.Context, uid uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
					captured = in
					ev, _ := domainevent.New(in.ID, uid, in.Title, "", in.StartAt, in.EndAt, "", "")
					return eventuc.UpdateEventOutput{Event: ev}, nil
				},
			}
			h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)
			w := httptest.NewRecorder()

			h.UpdateEvent(w, putEventRequestAsUser(t, uuid.NewString(), tt.body))

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}
			if !reflect.DeepEqual(captured.Recurrence, tt.wantInput.Recurrence) {
				t.Errorf("Recurrence: got %+v (spec %+v), want %+v (spec %+v)",
					captured.Recurrence, captured.Recurrence.Spec, tt.wantInput.Recurrence, tt.wantInput.Recurrence.Spec)
			}
			if captured.Scope != tt.wantInput.Scope {
				t.Errorf("Scope: got %q, want %q", captured.Scope, tt.wantInput.Scope)
			}
		})
	}
}

// TestEventHandler_UpdateEvent_SeriesPresent_ReturnsRecurringSeries covers a
// one-off event converted to a series.
func TestEventHandler_UpdateEvent_SeriesPresent_ReturnsRecurringSeries(t *testing.T) {
	userID := uuid.New()
	series := mustBuildWeeklyCount2Series(t, userID)
	stub := &stubUpdateEventExec{
		fn: func(_ context.Context, _ uuid.UUID, in eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			ev, _ := domainevent.New(in.ID, userID, in.Title, "", in.StartAt, in.EndAt, "", "")
			return eventuc.UpdateEventOutput{Event: ev.WithSeriesID(series.ID()), Series: series}, nil
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, putEventRequestAsUser(t, uuid.NewString(), validUpdateEventBody))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		RecurringSeries json.RawMessage `json:"recurringSeries"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	assertJSONEqual(t, resp.RecurringSeries, `{"id":"`+series.ID().String()+
		`","pattern":{"type":"weekly"},"end":{"type":"count","count":2},"timezone":"UTC"}`)
}

// TestEventHandler_UpdateEvent_ScopeValidationError_Returns400WithDetails
// verifies that a scope error from the command reaches the response details.
func TestEventHandler_UpdateEvent_ScopeValidationError_Returns400WithDetails(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(context.Context, uuid.UUID, eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{}, &domain.ValidationError{Details: []domain.ValidationDetail{
				{Field: "scope", Code: "SCOPE_NOT_ALLOWED", Message: "not allowed"},
			}}
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, putEventRequestAsUser(t, uuid.NewString(), validUpdateEventBody))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Details []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"details"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Details) != 1 || resp.Details[0].Field != "scope" || resp.Details[0].Code != "SCOPE_NOT_ALLOWED" {
		t.Errorf("details: got %+v", resp.Details)
	}
}

// The domain error carries no field, so the handler names startAt for the
// form to show the error next to it.
func TestEventHandler_UpdateEvent_DateChangeNotAllowed_Returns400WithStartAtDetail(t *testing.T) {
	stub := &stubUpdateEventExec{
		fn: func(context.Context, uuid.UUID, eventuc.UpdateEventInput) (eventuc.UpdateEventOutput, error) {
			return eventuc.UpdateEventOutput{}, fmt.Errorf("update: %w", rs.ErrRecurringOccurrenceDateChangeNotAllowed)
		},
	}
	h := handler.NewEventHandler(&stubListEventsExec{}, &stubCreateEventExec{}, stub, &stubDeleteEventExec{}, testLogger)
	w := httptest.NewRecorder()

	h.UpdateEvent(w, putEventRequestAsUser(t, uuid.NewString(), validUpdateEventBody))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code    string `json:"code"`
		Details []struct {
			Field string `json:"field"`
			Code  string `json:"code"`
		} `json:"details"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Code != "VALIDATION_ERROR" || len(resp.Details) != 1 ||
		resp.Details[0].Field != "startAt" || resp.Details[0].Code != "DATE_CHANGE_NOT_ALLOWED" {
		t.Errorf("got code %q details %+v", resp.Code, resp.Details)
	}
}
