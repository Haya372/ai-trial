package tag_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	taguc "github.com/Haya372/ai-trial/backend/usecase/tag"
)

type stubQueryService struct {
	fn func(context.Context, uuid.UUID) ([]taguc.TagReadModel, error)
}

func (s *stubQueryService) ListByUserID(ctx context.Context, userID uuid.UUID) ([]taguc.TagReadModel, error) {
	return s.fn(ctx, userID)
}

func TestListTagsQuery_Execute_ReturnsTagsOfUser(t *testing.T) {
	userID := uuid.New()
	want := []taguc.TagReadModel{
		{ID: uuid.New(), Name: "work", Color: "red", EventCount: 3},
		{ID: uuid.New(), Name: "private", Color: otherColor, EventCount: 0},
	}
	stub := &stubQueryService{
		fn: func(_ context.Context, gotUserID uuid.UUID) ([]taguc.TagReadModel, error) {
			if gotUserID != userID {
				t.Errorf("userID mismatch: got %v, want %v", gotUserID, userID)
			}
			return want, nil
		},
	}

	got, err := taguc.NewListTagsQuery(stub).Execute(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d tags, got %d", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("tags[%d] mismatch: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestListTagsQuery_Execute_EmptyResult(t *testing.T) {
	stub := &stubQueryService{
		fn: func(context.Context, uuid.UUID) ([]taguc.TagReadModel, error) {
			return []taguc.TagReadModel{}, nil
		},
	}

	got, err := taguc.NewListTagsQuery(stub).Execute(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected empty non-nil slice, got %#v", got)
	}
}

func TestListTagsQuery_Execute_QueryError(t *testing.T) {
	stub := &stubQueryService{
		fn: func(context.Context, uuid.UUID) ([]taguc.TagReadModel, error) {
			return nil, errDBFailure
		},
	}

	_, err := taguc.NewListTagsQuery(stub).Execute(context.Background(), uuid.New())
	if !errors.Is(err, errDBFailure) {
		t.Errorf("expected wrapped errDBFailure, got %v", err)
	}
}
