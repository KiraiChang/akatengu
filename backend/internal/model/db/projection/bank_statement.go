package projection

import "github.com/shopspring/decimal"

//dbmap:sqlcdb=BankCsvTemplate
type BankCsvTemplate struct {
	TemplateID        int64   `db:"template_id" json:"template_id"`
	TemplateUUID      string  `db:"template_uuid" json:"template_uuid"`
	MerchantID        int64   `db:"merchant_id" json:"merchant_id"`
	TemplateName      string  `db:"template_name" json:"template_name"`
	Encoding          string  `db:"encoding" json:"encoding"`
	SkipRows          int64   `db:"skip_rows" json:"skip_rows"`
	DateColumn        int64   `db:"date_column" json:"date_column"`
	DateFormat        string  `db:"date_format" json:"date_format"`
	DescriptionColumn int64   `db:"description_column" json:"description_column"`
	DebitColumn       *int64  `db:"debit_column" json:"debit_column"`
	CreditColumn      *int64  `db:"credit_column" json:"credit_column"`
	AmountColumn      *int64  `db:"amount_column" json:"amount_column"`
	BalanceColumn     *int64  `db:"balance_column" json:"balance_column"`
	ReferenceColumn   *int64  `db:"reference_column" json:"reference_column"`
	IsActive          bool    `db:"is_active" json:"is_active"`
	Note              *string `db:"note" json:"note"`
	Version           int64   `db:"version" json:"version"`
	UpdatedBy         *string `db:"updated_by" json:"updated_by"`
	UpdatedAt         string  `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=BankStatementImport
type BankStatementImport struct {
	ImportID      int64   `db:"import_id" json:"import_id"`
	ImportUUID    string  `db:"import_uuid" json:"import_uuid"`
	MerchantID    int64   `db:"merchant_id" json:"merchant_id"`
	LedgerID      int64   `db:"ledger_id" json:"ledger_id"`
	TemplateID    *int64  `db:"template_id" json:"template_id"`
	StatementDate string  `db:"statement_date" json:"statement_date"`
	ImportSource  string  `db:"import_source" json:"import_source"`
	Filename      *string `db:"filename" json:"filename"`
	Status        string  `db:"status" json:"status"`
	Note          *string `db:"note" json:"note"`
	Version       int64   `db:"version" json:"version"`
	UpdatedBy     *string `db:"updated_by" json:"updated_by"`
	UpdatedAt     string  `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=BankStatementTxn
type BankStatementTxn struct {
	BankTxnID       int64               `db:"bank_txn_id" json:"bank_txn_id"`
	BankTxnUUID     string              `db:"bank_txn_uuid" json:"bank_txn_uuid"`
	ImportID        int64               `db:"import_id" json:"import_id"`
	MerchantID      int64               `db:"merchant_id" json:"merchant_id"`
	TxnDate         string              `db:"txn_date" json:"txn_date"`
	Description     string              `db:"description" json:"description"`
	Debit           decimal.Decimal     `db:"debit" json:"debit"`
	Credit          decimal.Decimal     `db:"credit" json:"credit"`
	Balance         decimal.NullDecimal `db:"balance" json:"balance"`
	ReferenceNo     *string             `db:"reference_no" json:"reference_no"`
	MatchStatus     string              `db:"match_status" json:"match_status"`
	MatchedEntryID  *int64              `db:"matched_entry_id" json:"matched_entry_id"`
	MatchConfidence *string             `db:"match_confidence" json:"match_confidence"`
	CreatedTxnID    *int64              `db:"created_txn_id" json:"created_txn_id"`
}
