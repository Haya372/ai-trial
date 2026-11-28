package emailnotification

import (
	"maps"
	"slices"
)

// TemplateParams holds the values embedded into an EmailTemplate, such as a
// reset URL. They are kept only while the notification is pending
// (ADR-039) and must stay out of logs and traces.
type TemplateParams struct {
	values map[string]string
}

func NewTemplateParams(values map[string]string) TemplateParams {
	return TemplateParams{values: maps.Clone(values)}
}

func (p TemplateParams) Values() map[string]string {
	if p.values == nil {
		return map[string]string{}
	}
	return maps.Clone(p.values)
}

// Keys returns the keys in ascending order.
func (p TemplateParams) Keys() []string {
	return slices.Sorted(maps.Keys(p.values))
}

func (p TemplateParams) IsEmpty() bool { return len(p.values) == 0 }
