package process

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/runtime/mediator"
)

func Registry(m *mediator.Mediator) {
	m.Register(event_types.EventTransactionCFCategoryUpdated.Enum(), EventTransactionCFCategoryUpdatedHandle())
}
