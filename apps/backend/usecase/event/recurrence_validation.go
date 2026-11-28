package event

import (
	"errors"
	"fmt"
	"time"

	"github.com/Haya372/ai-trial/backend/domain"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

// buildRule returns the first problem as a ValidationError whose field is the
// request path (SPEC-009 POST /events), so the form can show it inline.
func buildRule(in RecurrenceInput) (rs.RecurrenceRule, error) {
	rule, err := parseRule(in)
	if err != nil {
		return rs.RecurrenceRule{}, err
	}
	if err := checkNewRule(rule); err != nil {
		return rs.RecurrenceRule{}, err
	}
	return rule, nil
}

// parseRule checks only the format, so PUT can accept the stored rule sent
// back unchanged: a series may have a count of 1 after a "this and following"
// change. Empty pattern type, end type and timezone are checked before the
// domain constructors, which would report them as INVALID_FORMAT instead of
// REQUIRED.
func parseRule(in RecurrenceInput) (rs.RecurrenceRule, error) {
	pattern, err := buildPattern(in)
	if err != nil {
		return rs.RecurrenceRule{}, err
	}
	end, err := buildEnd(in)
	if err != nil {
		return rs.RecurrenceRule{}, err
	}
	tz, err := buildTimezone(in)
	if err != nil {
		return rs.RecurrenceRule{}, err
	}
	return rs.RecurrenceRule{Pattern: pattern, End: end, Timezone: tz}, nil
}

// checkNewRule applies the limits a rule entered or changed by the user must
// meet beyond its format. The limits that depend on the first occurrence are
// left to ExpandOccurrences.
func checkNewRule(rule rs.RecurrenceRule) error {
	if n, ok := rule.End.Count(); ok {
		if _, err := rs.NewCountEnd(n); err != nil {
			return toEndValidationError(err)
		}
	}
	return nil
}

func buildPattern(in RecurrenceInput) (rs.RecurrencePattern, error) {
	if in.PatternType == "" {
		return rs.RecurrencePattern{}, validationDetail(
			"recurrence.pattern.type", "REQUIRED", "recurrence pattern type is required",
		)
	}
	pattern, err := rs.NewRecurrencePattern(in.PatternType, in.MonthlyNth)
	if err != nil {
		return rs.RecurrencePattern{}, toPatternValidationError(err, in.PatternType)
	}
	return pattern, nil
}

func buildEnd(in RecurrenceInput) (rs.RecurrenceEnd, error) {
	if in.EndType == "" {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.type", "REQUIRED", "recurrence end type is required",
		)
	}
	switch rs.EndType(in.EndType) {
	case rs.EndIndefinite:
		return rs.NewIndefiniteEnd(), nil
	case rs.EndUntilDate:
		return buildUntilDateEnd(in)
	case rs.EndCount:
		return buildCountEnd(in)
	}
	return rs.RecurrenceEnd{}, validationDetail(
		"recurrence.end.type", "INVALID_FORMAT", "recurrence end type is invalid",
	)
}

func buildUntilDateEnd(in RecurrenceInput) (rs.RecurrenceEnd, error) {
	if in.UntilDate == nil {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.untilDate",
			"REQUIRED",
			"untilDate is required when end type is until_date",
		)
	}
	t, err := time.Parse("2006-01-02", *in.UntilDate)
	if err != nil {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.untilDate", "INVALID_FORMAT", "untilDate must be YYYY-MM-DD",
		)
	}
	// time.Parse already rejects days missing from the calendar, so an error
	// here would be a bug rather than bad input.
	ld, err := rs.NewLocalDate(t.Year(), t.Month(), t.Day())
	if err != nil {
		return rs.RecurrenceEnd{}, fmt.Errorf("untilDate %s: %w", *in.UntilDate, err)
	}
	return rs.NewUntilDateEnd(ld), nil
}

func buildCountEnd(in RecurrenceInput) (rs.RecurrenceEnd, error) {
	if in.Count == nil {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.count", "REQUIRED", "count is required when end type is count",
		)
	}
	end, err := rs.RestoreCountEnd(*in.Count)
	if err != nil {
		return rs.RecurrenceEnd{}, toEndValidationError(err)
	}
	return end, nil
}

func buildTimezone(in RecurrenceInput) (rs.SeriesTimezone, error) {
	if in.Timezone == "" {
		return rs.SeriesTimezone{}, validationDetail(
			"recurrence.timezone", "REQUIRED", "timezone is required",
		)
	}
	tz, err := rs.NewSeriesTimezone(in.Timezone)
	if err != nil {
		return rs.SeriesTimezone{}, validationDetail(
			"recurrence.timezone", "INVALID_FORMAT", "timezone must be a valid IANA time zone name",
		)
	}
	return tz, nil
}

// toExpandValidationError names the field each ExpandOccurrences error comes
// from; the domain errors only carry the code.
func toExpandValidationError(err error) error {
	var de *domain.DomainError
	if !errors.As(err, &de) {
		return err
	}
	switch {
	case errors.Is(de, rs.ErrRecurringPatternInvalid):
		// Here the nth is a valid value that the first occurrence's date
		// does not fall on.
		return validationDetail("recurrence.pattern.monthlyNth", de.Code(), de.Message())
	case errors.Is(de, rs.ErrRecurringEndInvalid):
		return validationDetail("recurrence.end.untilDate", de.Code(), de.Message())
	case errors.Is(de, rs.ErrRecurringSeriesSpanExceeded):
		return validationDetail("recurrence.end.count", de.Code(), de.Message())
	case errors.Is(de, rs.ErrRecurringSeriesTooFewOccurrences):
		return validationDetail("recurrence", de.Code(), de.Message())
	}
	return err
}

func toPatternValidationError(err error, patternType string) error {
	var de *domain.DomainError
	if !errors.As(err, &de) {
		return err
	}
	switch {
	case errors.Is(de, rs.ErrRecurringPatternMonthlyNthRequired):
		return validationDetail("recurrence.pattern.monthlyNth", de.Code(), de.Message())
	case errors.Is(de, rs.ErrRecurringPatternInvalid):
		// The domain uses one error for a bad type and a bad monthlyNth, so a
		// known type means the monthlyNth is at fault.
		switch rs.PatternType(patternType) {
		case rs.PatternDaily, rs.PatternWeekly, rs.PatternBiweekly,
			rs.PatternMonthlyDay, rs.PatternMonthlyNthWeekday:
			return validationDetail("recurrence.pattern.monthlyNth", de.Code(), de.Message())
		default:
			return validationDetail("recurrence.pattern.type", de.Code(), de.Message())
		}
	}
	return err
}

func toEndValidationError(err error) error {
	var de *domain.DomainError
	if !errors.As(err, &de) {
		return err
	}
	if errors.Is(de, rs.ErrRecurringEndInvalid) {
		return validationDetail("recurrence.end.count", de.Code(), de.Message())
	}
	return err
}

func validationDetail(field, code, message string) error {
	return &domain.ValidationError{Details: []domain.ValidationDetail{
		{Field: field, Code: code, Message: message},
	}}
}
