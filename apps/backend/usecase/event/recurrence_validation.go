package event

import (
	"errors"
	"time"

	"github.com/Haya372/ai-trial/backend/domain"
	rs "github.com/Haya372/ai-trial/backend/domain/recurringeventseries"
)

// buildRule validates RecurrenceInput in the order specified in the design
// document and returns a RecurrenceRule or a ValidationError.
func buildRule(in RecurrenceInput) (rs.RecurrenceRule, error) {
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

// buildPattern validates and constructs the RecurrencePattern part of the rule.
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

// buildEnd validates and constructs the RecurrenceEnd part of the rule.
func buildEnd(in RecurrenceInput) (rs.RecurrenceEnd, error) {
	if in.EndType == "" {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.type", "REQUIRED", "recurrence end type is required",
		)
	}
	switch rs.EndType(in.EndType) {
	case rs.EndIndefinite, rs.EndUntilDate, rs.EndCount:
		// valid
	default:
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.type", "INVALID_FORMAT", "recurrence end type is invalid",
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
	// unreachable: all EndType values handled above
	return rs.RecurrenceEnd{}, nil
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
	ld, err := rs.NewLocalDate(t.Year(), t.Month(), t.Day())
	if err != nil {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.untilDate", "INVALID_FORMAT", "untilDate is not a valid calendar date",
		)
	}
	return rs.NewUntilDateEnd(ld), nil
}

func buildCountEnd(in RecurrenceInput) (rs.RecurrenceEnd, error) {
	if in.Count == nil {
		return rs.RecurrenceEnd{}, validationDetail(
			"recurrence.end.count", "REQUIRED", "count is required when end type is count",
		)
	}
	end, err := rs.NewCountEnd(*in.Count)
	if err != nil {
		return rs.RecurrenceEnd{}, toEndValidationError(err)
	}
	return end, nil
}

// buildTimezone validates and constructs the SeriesTimezone part of the rule.
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

// toExpandValidationError maps errors from ExpandOccurrences to ValidationErrors
// per the design document.
func toExpandValidationError(err error) error {
	var de *domain.DomainError
	if !errors.As(err, &de) {
		return err
	}
	switch {
	case errors.Is(de, rs.ErrRecurringPatternInvalid):
		// monthlyNth does not apply to the first occurrence
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

// toPatternValidationError maps NewRecurrencePattern errors to ValidationErrors.
func toPatternValidationError(err error, patternType string) error {
	var de *domain.DomainError
	if !errors.As(err, &de) {
		return err
	}
	switch {
	case errors.Is(de, rs.ErrRecurringPatternMonthlyNthRequired):
		return validationDetail("recurrence.pattern.monthlyNth", de.Code(), de.Message())
	case errors.Is(de, rs.ErrRecurringPatternInvalid):
		// type out of range → field is pattern.type; monthlyNth issue → monthlyNth
		switch rs.PatternType(patternType) {
		case rs.PatternDaily, rs.PatternWeekly, rs.PatternBiweekly,
			rs.PatternMonthlyDay, rs.PatternMonthlyNthWeekday:
			// type is valid; it must be a monthlyNth issue
			return validationDetail("recurrence.pattern.monthlyNth", de.Code(), de.Message())
		default:
			return validationDetail("recurrence.pattern.type", de.Code(), de.Message())
		}
	}
	return err
}

// toEndValidationError maps NewCountEnd errors to ValidationErrors.
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
