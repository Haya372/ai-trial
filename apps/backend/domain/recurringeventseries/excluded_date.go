package recurringeventseries

// ExcludedDate is a date the pattern would produce but the series does not
// keep, as a local date in the series' SeriesTimezone.
type ExcludedDate struct {
	date LocalDate
}

func NewExcludedDate(d LocalDate) ExcludedDate {
	return ExcludedDate{date: d}
}

func (e ExcludedDate) Date() LocalDate { return e.date }
