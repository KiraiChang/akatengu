package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=Account
//dbmap:sqlcdb=CreateAccountParams
//dbmap:sqlcdb=UpdateAccountParams
//dbmap:sqlcdb=GetAccountsPagedRow
type Account struct {
	AccountId     string              `db:"account_id" json:"account_id"`
	ParentId      *string             `db:"parent_id" json:"parent_id"`
	Name          string              `db:"name" json:"name"`
	Type          enums.AccountType   `db:"type" json:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance" json:"normal_balance"`
	Currency      string              `db:"currency" json:"currency"`
	IsSummary     bool                `db:"is_summary" json:"is_summary"`
	IsActive      bool                `db:"is_active" json:"is_active"`
	Note          *string             `db:"note" json:"note"`
	Version       int64               `db:"version" json:"version"`
}

//dbmap:sqlcdb=LedgerAccount
//dbmap:sqlcdb=CreateLedgerAccountParams
//dbmap:sqlcdb=UpdateLedgerAccountParams
type LedgerAccount struct {
	LedgerId    int64                   `db:"ledger_id"`
	AccountId   string                  `db:"account_id"`
	Institution string                  `db:"institution"`
	Name        string                  `db:"name"`
	Type        enums.LedgerAccountType `db:"type"`
	AccountNo   *string                 `db:"account_no"`
	Currency    string                  `db:"currency"`
	CreditLimit decimal.NullDecimal     `db:"credit_limit"`
	BillingDay  *string                 `db:"billing_day"`
	DueDay      *string                 `db:"due_day"`
	IsActive    bool                    `db:"is_active"`
	Note        *string                 `db:"note"`
	Version     int64                   `db:"version"`
}
