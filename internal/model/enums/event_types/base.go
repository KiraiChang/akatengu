package event_types

import (
	"akatengu/internal/pkg/enumx"
	"slices"
)

type EventType = enumx.EnumVal[eventTypeVal]

type eventTypeVal string

var validEventTypes = slices.Concat(
	validTransactionEventTypes,
	validInstallmentEventTypes,
	validReconciliationEventTypes,
	validAccountEventTypes,
	validInvestmentEventType,
	validCloseEventTypes,
)

func GetValidEventTypes() []eventTypeVal {
	return validEventTypes
}

func ParseEventType(s string) (EventType, error) {
	return enumx.New(s, validEventTypes)
}
