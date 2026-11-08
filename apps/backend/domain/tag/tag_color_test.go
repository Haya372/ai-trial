package tag_test

import (
	"errors"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/tag"
)

func TestNewTagColor_presetColors(t *testing.T) {
	presets := []string{
		"red", "orange", "yellow", "green", "teal", "cyan",
		"blue", "indigo", "purple", "pink", "brown", "gray",
	}
	for _, s := range presets {
		c, err := tag.NewTagColor(s)
		if err != nil {
			t.Errorf("NewTagColor(%q): unexpected error: %v", s, err)
			continue
		}
		if string(c) != s {
			t.Errorf("NewTagColor(%q): got %q", s, c)
		}
	}
}

func TestNewTagColor_invalid(t *testing.T) {
	for _, s := range []string{"", "Red", "black", "#ff0000", " red"} {
		_, err := tag.NewTagColor(s)
		if !errors.Is(err, tag.ErrTagColorInvalid) {
			t.Errorf("NewTagColor(%q): got %v, want ErrTagColorInvalid", s, err)
		}
	}
}
