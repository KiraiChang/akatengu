package payload

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

// AccountCreatePayload 範例
//
//	{
//	  "account_id": "1110-01",
//	  "parent_id": "1110",
//	  "name": "玉山銀行",
//	  "type": "asset",
//	  "normal_balance": "debit",
//	  "is_summary": 0
//	}
type AccountCreatePayload struct {
	AccountId        string                  `json:"account_id"`
	ParentId         *string                 `json:"parent_id"`
	Name             string                  `json:"name"`
	Type             enums.AccountType       `json:"type"`
	NormalBalance    enums.NormalBalance     `json:"normal_balance"`
	Currency         string                  `json:"currency"`
	IsSummary        bool                    `json:"is_summary"`
	IsActive         bool                    `json:"is_active"`
	Note             *string                 `json:"note"`
	CashFlowCategory *enums.CashFlowCategory `json:"cash_flow_category,omitempty"`
}

func (p AccountCreatePayload) Validate() error {
	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "account name is required")
	}

	return joinErrors(errs)
}

// AccountUpdatedPayload 範例
//
//	{
//	  "account_id": "1110-01",
//	  "parent_id": "1110",
//	  "name": "玉山銀行",
//	  "type": "asset",
//	  "normal_balance": "debit",
//	  "is_summary": 0
//	}
type AccountUpdatedPayload struct {
	AccountId        string                  `json:"account_id"`
	ParentId         *string                 `json:"parent_id"`
	Name             string                  `json:"name"`
	Type             enums.AccountType       `json:"type"`
	NormalBalance    enums.NormalBalance     `json:"normal_balance"`
	Currency         string                  `json:"currency"`
	IsSummary        bool                    `json:"is_summary"`
	IsActive         bool                    `json:"is_active"`
	Note             *string                 `json:"note"`
	Version          int64                   `json:"version"`
	CashFlowCategory *enums.CashFlowCategory `json:"cash_flow_category,omitempty"`
}

func (p AccountUpdatedPayload) Validate() error {
	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "account name is required")
	}

	if p.Version <= 0 {
		errs = append(errs, "version is required")
	}

	return joinErrors(errs)
}

// LedgerAccountCreatePayload 範例
//
//	{
//	  "account_id": "1110-01",
//	  "institution": "玉山銀行",
//	  "name": "玉山數位帳戶",
//	  "account_no": "1234",
//	  "currency": "TWD"
//	}
type LedgerAccountCreatePayload struct {
	AccountId   string                  `json:"account_id"`
	Institution string                  `json:"institution"`
	Name        string                  `json:"name"`
	AccountNo   *string                 `json:"account_no"`
	Currency    string                  `json:"currency"`
	CreditLimit decimal.NullDecimal     `json:"credit_limit"`
	BillingDay  *string                 `json:"billing_day"`
	DueDay      *string                 `json:"due_day"`
	IsActive    bool                    `json:"is_active"`
	Note        *string                 `json:"note"`
	Type        enums.LedgerAccountType `json:"type"`
}

func (p LedgerAccountCreatePayload) Validate() error {
	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}

	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	if p.Currency == "" {
		errs = append(errs, "currency is required")
	}

	if p.Institution == "" {
		errs = append(errs, "institution is required")
	}

	return joinErrors(errs)
}

// LedgerAccountUpdatedPayload 範例
//
//	{
//	  "ledger_id": 1,
//	  "account_id": "1110-01",
//	  "institution": "玉山銀行",
//	  "name": "玉山數位帳戶",
//	  "account_no": "1234",
//	  "currency": "TWD"
//	}
type LedgerAccountUpdatedPayload struct {
	LedgerId    int64                   `json:"ledger_id"`
	AccountId   string                  `json:"account_id"`
	Institution string                  `json:"institution"`
	Name        string                  `json:"name"`
	AccountNo   *string                 `json:"account_no"`
	Currency    string                  `json:"currency"`
	CreditLimit decimal.NullDecimal     `json:"credit_limit"`
	BillingDay  *string                 `json:"billing_day"`
	DueDay      *string                 `json:"due_day"`
	IsActive    bool                    `json:"is_active"`
	Note        *string                 `json:"note"`
	Version     int64                   `json:"version"`
	Type        enums.LedgerAccountType `json:"type"`
}

func (p LedgerAccountUpdatedPayload) Validate() error {
	var errs []string

	if p.LedgerId <= 0 {
		errs = append(errs, "ledger_id is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}

	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	if p.Currency == "" {
		errs = append(errs, "currency is required")
	}

	if p.Institution == "" {
		errs = append(errs, "institution is required")
	}

	if p.Version <= 0 {
		errs = append(errs, "version is required")
	}

	return joinErrors(errs)
}
