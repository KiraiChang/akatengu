package projection

import "akatengu/internal/enums"

//dbmap:sqlcdb=EntryCfCategory
type EntryCFCategory struct {
	EntryUuid   string                 `db:"entry_uuid"   json:"entry_uuid"`
	MerchantID  int64                  `db:"merchant_id"  json:"merchant_id"`
	CfCategory  enums.CashFlowCategory `db:"cf_category"  json:"cf_category"`
	IsConfirmed bool                   `db:"is_confirmed" json:"is_confirmed"`
	UpdatedAt   *string                `db:"updated_at"   json:"updated_at"`
	UpdatedBy   *string                `db:"updated_by"   json:"updated_by"`
}
