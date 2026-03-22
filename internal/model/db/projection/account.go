package projection

import (
	"akatengu/internal/model/enums"

	"github.com/shopspring/decimal"
)

type Account struct {
	AccountId     string              `db:"account_id"`
	ParentId      *string             `db:"parent_id"`
	Name          string              `db:"name"`
	Type          enums.AccountType   `db:"type"`
	NormalBalance enums.NormalBalance `db:"normal_balance"`
	Currency      string              `db:"currency"`
	IsSummary     bool                `db:"is_summary"`
	IsActive      bool                `db:"is_active"`
	Note          *string             `db:"note"`
	Version       int64               `db:"version"`
}

type LedgerAccount struct {
	LedgerId    int64               `db:"ledger_id"`
	AccountId   string              `db:"account_id"`
	Institution string              `db:"institution"`
	Name        string              `db:"name"`
	AccountNo   *string             `db:"account_no"`
	Currency    string              `db:"currency"`
	CreditLimit decimal.NullDecimal `db:"credit_limit"`
	BillingDay  *string             `db:"billing_day"`
	DueDay      *string             `db:"due_day"`
	IsActive    bool                `db:"is_active"`
	Note        *string             `db:"note"`
	Version     int64               `db:"version"`
}
