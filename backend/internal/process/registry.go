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
	m.Register(event_types.EventLedgerAccountCreated.Enum(), EventLedgerAccountCreatedHandle())
	m.Register(event_types.EventLedgerAccountUpdated.Enum(), EventLedgerAccountUpdatedHandle())

	m.Register(event_types.EventSysAccountUpdated.Enum(), EventSysAccountUpdatedHandle())
	m.Register(event_types.EventLedgerAccountTypeConfigUpdated.Enum(), EventLedgerAccountTypeConfigUpdatedHandle())
	m.Register(event_types.EventAssetTypeAccountConfigUpdated.Enum(), EventAssetTypeAccountConfigUpdatedHandle())

	m.Register(event_types.EventPrepaidCategoryCreated.Enum(), validateHandle[event_payload.PrepaidCategoryCreatedPayload]())
	m.Register(event_types.EventPrepaidCategoryUpdated.Enum(), EventPrepaidCategoryUpdatedHandle())
	m.Register(event_types.EventPrepaidCategoryDeleted.Enum(), EventPrepaidCategoryDeletedHandle())
	m.Register(event_types.EventFixedAssetCategoryCreated.Enum(), validateHandle[event_payload.FixedAssetCategoryCreatedPayload]())
	m.Register(event_types.EventFixedAssetCategoryUpdated.Enum(), EventFixedAssetCategoryUpdatedHandle())
	m.Register(event_types.EventFixedAssetCategoryDeleted.Enum(), EventFixedAssetCategoryDeletedHandle())

	// Opening a period only validates the payload, as the legacy pipelines do
	// (services/pipelines/factory/period.go).
	m.Register(event_types.EventPeriodMonthStarted.Enum(), validateHandle[event_payload.PeriodMonthStartedPayload]())
	m.Register(event_types.EventPeriodAnnualStarted.Enum(), validateHandle[event_payload.PeriodAnnualStartedPayload]())
}
