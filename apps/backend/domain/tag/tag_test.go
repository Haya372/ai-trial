package tag_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/tag"
)

func TestNew(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()

	tg := tag.New(id, userID, tag.TagName("work"), tag.ColorBlue)

	if tg.ID() != id {
		t.Errorf("ID mismatch: got %v, want %v", tg.ID(), id)
	}
	if tg.UserID() != userID {
		t.Errorf("UserID mismatch: got %v, want %v", tg.UserID(), userID)
	}
	if tg.Name() != "work" {
		t.Errorf("Name mismatch: got %q, want %q", tg.Name(), "work")
	}
	if tg.Color() != tag.ColorBlue {
		t.Errorf("Color mismatch: got %q, want %q", tg.Color(), tag.ColorBlue)
	}
}
