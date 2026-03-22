package enums

import (
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/pkg/enumx"
)

func InitEnums() {
	enumx.Register(validAccountTypes)
	enumx.Register(validAggregateTypes)
	enumx.Register(event_types.GetValidEventTypes())
	enumx.Register(validNormalBalances)
	enumx.Register(validPeriodStatuses)
	enumx.Register(validTransactionStatuses)
	enumx.Register(validPeriodStatuses)
	enumx.Register(validClosingStatuses)
	enumx.Register(validAssetTypes)
	enumx.Register(validCostMethods)
	enumx.Register(validMovementTypes)
}
