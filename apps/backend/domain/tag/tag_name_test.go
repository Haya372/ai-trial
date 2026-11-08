package tag_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Haya372/ai-trial/backend/domain/tag"
)

func TestNewTagName_valid(t *testing.T) {
	name, err := tag.NewTagName("work")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "work" {
		t.Errorf("got %q, want %q", name, "work")
	}
}

func TestNewTagName_trimsSurroundingSpaces(t *testing.T) {
	name, err := tag.NewTagName("  remote work \t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "remote work" {
		t.Errorf("got %q, want %q", name, "remote work")
	}
}

func TestNewTagName_empty(t *testing.T) {
	for _, s := range []string{"", "   ", "\t\n"} {
		_, err := tag.NewTagName(s)
		if !errors.Is(err, tag.ErrTagNameEmpty) {
			t.Errorf("NewTagName(%q): got %v, want ErrTagNameEmpty", s, err)
		}
	}
}

func TestNewTagName_minLength(t *testing.T) {
	if _, err := tag.NewTagName("a"); err != nil {
		t.Errorf("1-char tag name should be valid, got: %v", err)
	}
}

func TestNewTagName_maxLengthCountsRunes(t *testing.T) {
	if _, err := tag.NewTagName(strings.Repeat("あ", 20)); err != nil {
		t.Errorf("20-char tag name should be valid, got: %v", err)
	}
}

func TestNewTagName_maxLengthAfterTrim(t *testing.T) {
	if _, err := tag.NewTagName(" " + strings.Repeat("a", 20) + " "); err != nil {
		t.Errorf("20-char tag name with surrounding spaces should be valid, got: %v", err)
	}
}

func TestNewTagName_tooLong(t *testing.T) {
	_, err := tag.NewTagName(strings.Repeat("a", 21))
	if !errors.Is(err, tag.ErrTagNameTooLong) {
		t.Errorf("got %v, want ErrTagNameTooLong", err)
	}
}
