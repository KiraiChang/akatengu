package event_types

import "akatengu/internal/pkg/enumx"

const (
	// account
	eventAccountCreated           eventTypeVal = "account.created"
	eventAccountUpdated           eventTypeVal = "account.updated"
	eventAccountDeactivated       eventTypeVal = "account.deactivated"
	eventLedgerAccountCreated     eventTypeVal = "ledger_account.created"
	eventLedgerAccountUpdated     eventTypeVal = "ledger_account.updated"
	eventLedgerAccountDeactivated eventTypeVal = "ledger_account.deactivated"
)

var validAccountEventTypes = []eventTypeVal{
	eventAccountCreated,
	eventAccountUpdated,
	eventAccountDeactivated,
	eventLedgerAccountCreated,
	eventLedgerAccountDeactivated,
	eventLedgerAccountCreated,
	eventLedgerAccountUpdated,
	eventLedgerAccountDeactivated,
}

var (
	// account
	EventAccountCreated           = enumx.Must(string(eventAccountCreated), validEventTypes)
	EventAccountUpdated           = enumx.Must(string(eventAccountUpdated), validEventTypes)
	EventAccountDeactivated       = enumx.Must(string(eventAccountDeactivated), validEventTypes)
	EventLedgerAccountCreated     = enumx.Must(string(eventLedgerAccountCreated), validEventTypes)
	EventLedgerAccountUpdated     = enumx.Must(string(eventLedgerAccountUpdated), validEventTypes)
	EventLedgerAccountDeactivated = enumx.Must(string(eventLedgerAccountDeactivated), validEventTypes)
)
