package event

import "akatengu/internal/enums/event_types"

// The period opening payloads are copies of model/payload/period.go (BFS R2
// D1); the JSON tags must stay identical until R10. Closing and reopening stay
// on the legacy path until R5.

// PeriodMonthStartedPayload opens the month PeriodStart ("2006-01").
type PeriodMonthStartedPayload struct {
	PeriodStart string `json:"period_start"`
}

func (p PeriodMonthStartedPayload) EventType() event_types.EventType {
	return event_types.EventPeriodMonthStarted.Enum()
}

func (p PeriodMonthStartedPayload) Validate() error {
	var errs []string
	if p.PeriodStart == "" {
		errs = append(errs, "period_start is required")
	}
	return joinErrors(errs)
}

// PeriodAnnualStartedPayload opens the year Year.
type PeriodAnnualStartedPayload struct {
	Year int `json:"year"`
}

func (p PeriodAnnualStartedPayload) EventType() event_types.EventType {
	return event_types.EventPeriodAnnualStarted.Enum()
}

func (p PeriodAnnualStartedPayload) Validate() error {
	var errs []string
	if p.Year == 0 {
		errs = append(errs, "year is required")
	}
	return joinErrors(errs)
}
