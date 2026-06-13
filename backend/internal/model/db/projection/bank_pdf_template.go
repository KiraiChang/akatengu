package projection

//dbmap:sqlcdb=BankPdfTemplate
type BankPdfTemplate struct {
	TemplateID   int64   `db:"template_id"   json:"template_id"`
	TemplateUUID string  `db:"template_uuid" json:"template_uuid"`
	MerchantID   int64   `db:"merchant_id"   json:"merchant_id"`
	TemplateName string  `db:"template_name" json:"template_name"`
	BankType     string  `db:"bank_type"     json:"bank_type"`
	IsActive     bool    `db:"is_active"     json:"is_active"`
	Version      int64   `db:"version"       json:"version"`
	UpdatedBy    *string `db:"updated_by"    json:"updated_by"`
	UpdatedAt    string  `db:"updated_at"    json:"updated_at"`
}

//dbmap:sqlcdb=BankPdfTemplateLedger
type BankPdfTemplateLedger struct {
	TplLedgerID   int64   `db:"tpl_ledger_id"   json:"tpl_ledger_id"`
	TplLedgerUUID string  `db:"tpl_ledger_uuid" json:"tpl_ledger_uuid"`
	TemplateID    int64   `db:"template_id"     json:"template_id"`
	LedgerUUID    string  `db:"ledger_uuid"     json:"ledger_uuid"`
	LedgerID      *int64  `db:"ledger_id"       json:"ledger_id"`
	AccountType   string  `db:"account_type"    json:"account_type"`
	SortOrder     int64   `db:"sort_order"      json:"sort_order"`
}

//dbmap:sqlcdb=BankCsvTemplateLedger
type BankCsvTemplateLedger struct {
	TplLedgerID   int64   `db:"tpl_ledger_id"   json:"tpl_ledger_id"`
	TplLedgerUUID string  `db:"tpl_ledger_uuid" json:"tpl_ledger_uuid"`
	TemplateID    int64   `db:"template_id"     json:"template_id"`
	LedgerUUID    string  `db:"ledger_uuid"     json:"ledger_uuid"`
	LedgerID      *int64  `db:"ledger_id"       json:"ledger_id"`
	AccountType   string  `db:"account_type"    json:"account_type"`
	SortOrder     int64   `db:"sort_order"      json:"sort_order"`
}

//dbmap:sqlcdb=BankStatementImportLedger
type BankStatementImportLedger struct {
	ImportLedgerID   int64  `db:"import_ledger_id"   json:"import_ledger_id"`
	ImportLedgerUUID string `db:"import_ledger_uuid" json:"import_ledger_uuid"`
	ImportID         int64  `db:"import_id"          json:"import_id"`
	LedgerUUID       string `db:"ledger_uuid"        json:"ledger_uuid"`
	LedgerID         *int64 `db:"ledger_id"          json:"ledger_id"`
	AccountType      string `db:"account_type"       json:"account_type"`
}
