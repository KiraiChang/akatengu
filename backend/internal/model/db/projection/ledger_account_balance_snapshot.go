package projection

import "github.com/shopspring/decimal"

type LedgerAccountBalanceSnapshot struct {
	ClosingId   int64           `db:"closing_id"`
	LedgerId    int64           `db:"ledger_id"`
	DebitTotal  decimal.Decimal `db:"debit_total"`
	CreditTotal decimal.Decimal `db:"credit_total"`
}
