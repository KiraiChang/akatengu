package projection

import "github.com/shopspring/decimal"

// AccountChildBalance is a composite type (manual mapping, no dbmap annotation).
type AccountChildBalance struct {
	AccountID     string          `db:"account_id" json:"account_id"`
	ParentID      *string         `db:"parent_id" json:"parent_id"`
	Name          string          `db:"name" json:"name"`
	IsSummary     bool            `db:"is_summary" json:"is_summary"`
	HasChild      bool            `db:"has_child" json:"has_child"`
	DebitTotal    decimal.Decimal `json:"debit_total"`
	CreditTotal   decimal.Decimal `json:"credit_total"`
}

type AccountChildrenBalanceResult struct {
	AccountID   string                `json:"account_id"`
	Name        string                `json:"name"`
	Children    []AccountChildBalance `json:"children"`
}

// AccountJournalEntryRow is a composite type (manual mapping, no dbmap annotation).
type AccountJournalEntryRow struct {
	EntryID     int64           `db:"entry_id" json:"entry_id"`
	TxnID       int64           `db:"txn_id" json:"txn_id"`
	AccountID   string          `db:"account_id" json:"account_id"`
	LedgerID    *int64          `db:"ledger_id" json:"ledger_id"`
	Debit       decimal.Decimal `db:"debit" json:"debit"`
	Credit      decimal.Decimal `db:"credit" json:"credit"`
	Note        *string         `db:"note" json:"note"`
	TxnDate     string          `db:"txn_date" json:"txn_date"`
	Description string          `db:"description" json:"description"`
}

// AccountMonthlyBalance is computed in Go from snapshot + delta queries.
type AccountMonthlyBalance struct {
	Month       string          `json:"month"`
	DebitTotal  decimal.Decimal `json:"debit_total"`
	CreditTotal decimal.Decimal `json:"credit_total"`
	HasSnapshot bool            `json:"has_snapshot"`
}
