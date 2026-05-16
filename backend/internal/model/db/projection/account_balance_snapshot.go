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
//
//dbmap:sqlcdb=GetAllAccountBalancesRow
//dbmap:sqlcdb=GetAllAccountRunningBalanceRow
type AccountBalance struct {
	AccountId     string              `db:"account_id" json:"account_id"`
	Name          string              `db:"name" json:"name"`
	Type          enums.AccountType   `db:"type" json:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance" json:"normal_balance"`
	DebitTotal    decimal.Decimal     `db:"debit_total" json:"debit_total"`
	CreditTotal   decimal.Decimal     `db:"credit_total" json:"credit_total"`
}
