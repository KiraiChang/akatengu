package event_types

import (
	"akatengu/internal/pkg/enumx"
)

const (
	// month closing.*
	eventPeriodMonthStarted  eventTypeVal = "period.month_started"
	eventPeriodMonthClosed   eventTypeVal = "period.month_closed"
	eventPeriodMonthReopened eventTypeVal = "period.month_reopened"

	// annual closing.*
	eventPeriodAnnualStarted  eventTypeVal = "period.annual_started"
	eventPeriodAnnualClosed   eventTypeVal = "period.annual_closed"
	eventPeriodAnnualReopened eventTypeVal = "period.annual_reopened"
)

var validCloseEventTypes = []eventTypeVal{
	// month closing.*
	eventPeriodMonthStarted,
	eventPeriodMonthClosed,
	eventPeriodMonthReopened,

	// annual closing.*
	eventPeriodAnnualStarted,
	eventPeriodAnnualClosed,
	eventPeriodAnnualReopened,
}

var (
	// month closing.*

	EventPeriodMonthStarted  = enumx.Must(string(eventPeriodMonthStarted), validEventTypes)
	EventPeriodMonthClosed   = enumx.Must(string(eventPeriodMonthClosed), validEventTypes)
	EventPeriodMonthReopened = enumx.Must(string(eventPeriodMonthReopened), validEventTypes)

	// annual closing.*

	EventPeriodAnnualStarted  = enumx.Must(string(eventPeriodAnnualStarted), validEventTypes)
	EventPeriodAnnualClosed   = enumx.Must(string(eventPeriodAnnualClosed), validEventTypes)
	EventPeriodAnnualReopened = enumx.Must(string(eventPeriodAnnualReopened), validEventTypes)
)
