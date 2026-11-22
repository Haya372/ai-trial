package recurringeventseries

import (
	"errors"

	"github.com/Haya372/ai-trial/backend/domain"
)

// The codes double as details[].code of the recurrence validation error
// (SPEC-009 POST /events), so several errors share the same code.
const (
	CodeRecurringPatternInvalid          = "INVALID_FORMAT"
	CodeRecurringEndInvalid              = "OUT_OF_RANGE"
	CodeRecurringTimezoneInvalid         = "INVALID_FORMAT"
	CodeRecurringSeriesTooFewOccurrences = "TOO_FEW_OCCURRENCES"
	CodeRecurringSeriesSpanExceeded      = "OUT_OF_RANGE"
	CodeLocalDateInvalid                 = "INVALID_FORMAT"
)

var (
	ErrRecurringPatternInvalid = domain.NewDomainError(
		CodeRecurringPatternInvalid, "Recurrence pattern is invalid")
	ErrRecurringEndInvalid = domain.NewDomainError(
		CodeRecurringEndInvalid, "Recurrence end is out of range")
	ErrRecurringTimezoneInvalid = domain.NewDomainError(
		CodeRecurringTimezoneInvalid, "Timezone must be an IANA time zone name")
	ErrRecurringSeriesTooFewOccurrences = domain.NewDomainError(
		CodeRecurringSeriesTooFewOccurrences, "A recurring series must have at least 2 occurrences")
	ErrRecurringSeriesSpanExceeded = domain.NewDomainError(
		CodeRecurringSeriesSpanExceeded, "Occurrences must start within 2 years of the first occurrence")
	ErrLocalDateInvalid = domain.NewDomainError(
		CodeLocalDateInvalid, "Date does not exist in the calendar")
)

// ErrRecurrenceRuleUninitialized is not a DomainError: a zero-value rule
// comes from a caller that skipped the constructors, not from user input, so
// it surfaces as an internal error instead of a 400.
var ErrRecurrenceRuleUninitialized = errors.New("recurrence rule has an uninitialized pattern, end or timezone")
