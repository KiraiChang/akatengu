package process

import (
	event_payload "akatengu/internal/domain/event"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/runtime/mediator"
)

func Registry(m *mediator.Mediator) {
	m.Register(event_types.EventTransactionCFCategoryUpdated.Enum(), EventTransactionCFCategoryUpdatedHandle())

	m.Register(event_types.EventAccountCreated.Enum(), validateHandle[event_payload.AccountCreatePayload]())
	m.Register(event_types.EventAccountUpdated.Enum(), EventAccountUpdatedHandle())
	m.Register(event_types.EventAccountBulkImported.Enum(), validateHandle[event_payload.AccountBulkImportedPayload]())
	m.Register(event_types.EventLedgerAccountCreated.Enum(), validateHandle[event_payload.LedgerAccountCreatePayload]())
	m.Register(event_types.EventLedgerAccountUpdated.Enum(), EventLedgerAccountUpdatedHandle())
}
