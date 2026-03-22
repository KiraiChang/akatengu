package event_types

import (
	"akatengu/internal/pkg/enumx"
)

const (
	// closing.*
	eventPeriodCloseStarted eventTypeVal = "period.close_started"
	eventPeriodClosed       eventTypeVal = "period.closed"
	eventPeriodReopened     eventTypeVal = "period.reopened"

	// annual closing.*
	eventAnnualClosingEntryCreated eventTypeVal = "annual.closing_entry_created"
	eventAnnualOpeningEntryCreated eventTypeVal = "annual.opening_entry_created"
	eventAnnualClosingEntryVoided  eventTypeVal = "annual.closing_entry_voided"
	eventAnnualOpeningEntryVoided  eventTypeVal = "annual.opening_entry_voided"
)

var validCloseEventTypes = []eventTypeVal{
	// closing.*
	eventPeriodCloseStarted,
	eventPeriodClosed,
	eventPeriodReopened,

	// annual closing.*
	eventAnnualClosingEntryCreated,
	eventAnnualOpeningEntryCreated,
	eventAnnualClosingEntryVoided,
	eventAnnualOpeningEntryVoided,
}

var (
	// closing.*
	EventPeriodCloseStarted = enumx.Must(string(eventPeriodCloseStarted), validEventTypes)
	EventPeriodClosed       = enumx.Must(string(eventPeriodClosed), validEventTypes)
	EventPeriodReopened     = enumx.Must(string(eventPeriodReopened), validEventTypes)

	// annual closing.*
	EventAnnualClosingEntryCreated = enumx.Must(string(eventAnnualClosingEntryCreated), validEventTypes)
	EventAnnualOpeningEntryCreated = enumx.Must(string(eventAnnualOpeningEntryCreated), validEventTypes)
	EventAnnualClosingEntryVoided  = enumx.Must(string(eventAnnualClosingEntryVoided), validEventTypes)
	EventAnnualOpeningEntryVoided  = enumx.Must(string(eventAnnualOpeningEntryVoided), validEventTypes)
)
