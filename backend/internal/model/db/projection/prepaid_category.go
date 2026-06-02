package projection

//dbmap:sqlcdb=PrepaidCategory
type PrepaidCategory struct {
	ID               int64   `db:"id" json:"id"`
	CategoryUUID     string  `db:"category_uuid" json:"category_uuid"`
	MerchantID       int64   `db:"merchant_id" json:"merchant_id"`
	Name             string  `db:"name" json:"name"`
	AccountID        string  `db:"account_id" json:"account_id"`
	ExpenseAccountID string  `db:"expense_account_id" json:"expense_account_id"`
	IsActive         bool    `db:"is_active" json:"is_active"`
	UpdatedBy        *string `db:"updated_by" json:"updated_by"`
	UpdatedAt        string  `db:"updated_at" json:"updated_at"`
	Version          int64   `db:"version" json:"version"`
}
