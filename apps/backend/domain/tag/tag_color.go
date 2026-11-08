package tag

import (
	"fmt"

	"github.com/Haya372/ai-trial/backend/domain"
)

// TagColor is a preset color key. The display color (HEX) for each key lives
// only in the frontend theme, so the palette can be retuned without migrating
// stored data (SPEC-008). The keys mirror the OpenAPI TagColor enum (ADR-019).
type TagColor string

const (
	ColorRed    TagColor = "red"
	ColorOrange TagColor = "orange"
	ColorYellow TagColor = "yellow"
	ColorGreen  TagColor = "green"
	ColorTeal   TagColor = "teal"
	ColorCyan   TagColor = "cyan"
	ColorBlue   TagColor = "blue"
	ColorIndigo TagColor = "indigo"
	ColorPurple TagColor = "purple"
	ColorPink   TagColor = "pink"
	ColorBrown  TagColor = "brown"
	ColorGray   TagColor = "gray"
)

const CodeTagColorInvalid = "INVALID_FORMAT"

var ErrTagColorInvalid = domain.NewDomainError(CodeTagColorInvalid, "Tag color must be one of the preset colors")

var presetColors = map[TagColor]struct{}{
	ColorRed: {}, ColorOrange: {}, ColorYellow: {}, ColorGreen: {},
	ColorTeal: {}, ColorCyan: {}, ColorBlue: {}, ColorIndigo: {},
	ColorPurple: {}, ColorPink: {}, ColorBrown: {}, ColorGray: {},
}

func NewTagColor(s string) (TagColor, error) {
	c := TagColor(s)
	if _, ok := presetColors[c]; !ok {
		return "", fmt.Errorf("%w: %q", ErrTagColorInvalid, s)
	}
	return c, nil
}
