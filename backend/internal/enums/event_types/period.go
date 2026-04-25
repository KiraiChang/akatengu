package event_types

const (
	// month closing.*
	EventPeriodMonthStarted  eventTypeVal = "period.month_started"
	EventPeriodMonthClosed   eventTypeVal = "period.month_closed"
	EventPeriodMonthReopened eventTypeVal = "period.month_reopened"

	// annual closing.*
	EventPeriodAnnualStarted  eventTypeVal = "period.annual_started"
	EventPeriodAnnualClosed   eventTypeVal = "period.annual_closed"
	EventPeriodAnnualReopened eventTypeVal = "period.annual_reopened"
)
