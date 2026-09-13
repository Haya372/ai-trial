package event_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/Haya372/ai-trial/backend/domain/event"
	mock "github.com/Haya372/ai-trial/backend/domain/event/generated"
	eventuc "github.com/Haya372/ai-trial/backend/usecase/event"
)

func makeEvent(t *testing.T, title string, start, end time.Time) event.Event {
	t.Helper()
	e, err := event.New(uuid.New(), uuid.New(), title, "", start, end)
	if err != nil {
		t.Fatalf("make event: %v", err)
	}
	return e
}

func TestListEventsQuery_Execute_ReturnsEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockQueryRepository(ctrl)

	userID := uuid.New()
	now := time.Now()
	start := now
	end := now.Add(24 * time.Hour)

	ev := makeEvent(t, "Team meeting", start, start.Add(time.Hour))
	repo.EXPECT().
		List(gomock.Any(), event.ListFilter{UserID: userID, StartDate: start, EndDate: end}).
		Return([]event.Event{ev}, nil)

	q := eventuc.NewListEventsQuery(repo)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 event, got %d", len(result))
	}
	if result[0].Title != "Team meeting" {
		t.Errorf("title mismatch: got %q", result[0].Title)
	}
}

func TestListEventsQuery_Execute_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockQueryRepository(ctrl)

	userID := uuid.New()
	now := time.Now()
	repoErr := errDBFailure

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return(nil, repoErr)

	q := eventuc.NewListEventsQuery(repo)
	_, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListEventsQuery_Execute_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mock.NewMockQueryRepository(ctrl)

	userID := uuid.New()
	now := time.Now()

	repo.EXPECT().
		List(gomock.Any(), gomock.Any()).
		Return([]event.Event{}, nil)

	q := eventuc.NewListEventsQuery(repo)
	result, err := q.Execute(context.Background(), userID, eventuc.ListEventsInput{
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d events", len(result))
	}
}
