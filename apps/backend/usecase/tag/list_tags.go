package tag

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type TagReadModel struct {
	ID    uuid.UUID
	Name  string
	Color string
	// EventCount is the number of events the tag is attached to. The client
	// shows it in the delete confirmation dialog (SPEC-008).
	EventCount int
}

type ListTagsQuery struct {
	queryService QueryService
}

func NewListTagsQuery(s QueryService) *ListTagsQuery {
	return &ListTagsQuery{queryService: s}
}

func (q *ListTagsQuery) Execute(ctx context.Context, userID uuid.UUID) ([]TagReadModel, error) {
	tags, err := q.queryService.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}
