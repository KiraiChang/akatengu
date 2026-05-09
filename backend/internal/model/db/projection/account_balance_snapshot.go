package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type AccountBalanceSnapshot struct {
	ClosingId   int64           `db:"closing_id"`
	AccountId   string          `db:"account_id"`
	DebitTotal  decimal.Decimal `db:"debit_total"`
	CreditTotal decimal.Decimal `db:"credit_total"`
}

// AccountBalance 是 GetAccountSummaryOptimized 的查詢結果，包含科目餘額明細
type AccountBalance struct {
	AccountId     string              `db:"account_id"`
	Name          string              `db:"name"`
	Type          enums.AccountType   `db:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance"`
	DebitTotal    decimal.Decimal     `db:"debit_total"`
	CreditTotal   decimal.Decimal     `db:"credit_total"`
}