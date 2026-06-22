package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"fmt"
)

// TransactionCFEntryItem represents a single entry's CF category update
type TransactionCFEntryItem struct {
	EntryUUID  string                 `json:"entry_uuid"`
	CFCategory enums.CashFlowCategory `json:"cf_category"`
}

// TransactionCFCategoryUpdatedPayload is the payload for transaction.cf_category_updated
type TransactionCFCategoryUpdatedPayload struct {
	TxnUUID string                   `json:"txn_uuid"`
	Entries []TransactionCFEntryItem `json:"entries"`
}

func (p TransactionCFCategoryUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventTransactionCFCategoryUpdated.Enum()
}

func (p TransactionCFCategoryUpdatedPayload) Validate() error {
	var errs []string
	if p.TxnUUID == "" {
		errs = append(errs, "txn_uuid is required")
	}
	if len(p.Entries) == 0 {
		errs = append(errs, "entries must not be empty")
	}
	for i, e := range p.Entries {
		if e.EntryUUID == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].entry_uuid is required", i))
		}
		if !e.CFCategory.In(enums.CashFlowCategoryOperating, enums.CashFlowCategoryInvesting, enums.CashFlowCategoryFinancing) {
			errs = append(errs, fmt.Sprintf("entries[%d].cf_category must be OPERATING, INVESTING, or FINANCING", i))
		}
	}
	return joinErrors(errs)
}
