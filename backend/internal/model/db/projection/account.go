package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=Account
//dbmap:sqlcdb=CreateAccountParams
//dbmap:sqlcdb=UpdateAccountParams
//dbmap:sqlcdb=UpsertAccountParams
//dbmap:sqlcdb=GetAccountPagedRow
//dbmap:sqlcdb=GetAccountRow
//dbmap:sqlcdb=GetChildrenAccountRow
//dbmap:sqlcdb=GetAllAccountsRow
type Account struct {
	MerchantID       int64                   `db:"merchant_id" json:"merchant_id"`
	AccountId        string                  `db:"account_id" json:"account_id"`
	ParentId         *string                 `db:"parent_id" json:"parent_id"`
	Name             string                  `db:"name" json:"name"`
	Type             enums.AccountType       `db:"type" json:"type"`
	NormalBalance    enums.NormalBalance     `db:"normal_balance" json:"normal_balance"`
	Currency         string                  `db:"currency" json:"currency"`
	IsSummary        bool                    `db:"is_summary" json:"is_summary"`
	IsActive         bool                    `db:"is_active" json:"is_active"`
	Note             *string                 `db:"note" json:"note"`
	Version          int64                   `db:"version" json:"version"`
	HasChild         bool                    `db:"has_child" json:"has_child"`
	CashFlowCategory *enums.CashFlowCategory `db:"cash_flow_category" json:"cash_flow_category"`
	UpdatedBy        *string                 `db:"updated_by" json:"updated_by"`
	UpdatedAt        *string                 `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=LedgerAccount
//dbmap:sqlcdb=CreateLedgerAccountParams
//dbmap:sqlcdb=UpdateLedgerAccountParams
//dbmap:sqlcdb=GetLedgerPagedRow
//dbmap:sqlcdb=GetLedgerRow
//dbmap:sqlcdb=GetAllLedgersRow
type LedgerAccount struct {
	MerchantID  int64                   `db:"merchant_id" json:"merchant_id"`
	LedgerId    int64                   `db:"ledger_id" json:"ledger_id"`
	AccountId   string                  `db:"account_id" json:"account_id"`
	Institution string                  `db:"institution" json:"institution"`
	Name        string                  `db:"name" json:"name"`
	Type        enums.LedgerAccountType `db:"type" json:"type"`
	AccountNo   *string                 `db:"account_no" json:"account_no"`
	Currency    string                  `db:"currency" json:"currency"`
	CreditLimit decimal.NullDecimal     `db:"credit_limit" json:"credit_limit"`
	BillingDay  *string                 `db:"billing_day" json:"billing_day"`
	DueDay      *string                 `db:"due_day" json:"due_day"`
	IsActive    bool                    `db:"is_active" json:"is_active"`
	Note        *string                 `db:"note" json:"note"`
	Version     int64                   `db:"version" json:"version"`
	UpdatedBy   *string                 `db:"updated_by" json:"updated_by"`
	UpdatedAt   *string                 `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=GetAllLedgerBalancesRow
//dbmap:sqlcdb=GetAllLedgerRunningBalanceRow
type LedgerAccountBalance struct {
	LedgerId      int64               `db:"ledger_id" json:"ledger_id"`
	DebitTotal    decimal.Decimal     `db:"debit_total" json:"debit_total"`
	CreditTotal   decimal.Decimal     `db:"credit_total" json:"credit_total"`
	NormalBalance enums.NormalBalance `db:"normal_balance" json:"normal_balance"`
	Balance       decimal.Decimal     `db:"balance" json:"balance"`
}
