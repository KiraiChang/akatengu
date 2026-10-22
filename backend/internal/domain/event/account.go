package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"

	"github.com/shopspring/decimal"
)

// The account payloads are copies of model/payload/account.go (BFS R2 D1):
// the legacy pipeline, projection and replay keep using those, so the JSON tags
// must stay identical until R10 removes them.

// MerchantAggregateID is the aggregate ID of the aggregates a merchant has only
// one of: the account seed ('ACCOUNT', "1"), SYS_CONFIG and ACCOUNT_CONFIG
// (BFS R2 D8). It is a constant, not the merchant ID; the aggregate key
// includes merchant_id.
const MerchantAggregateID = "1"

// AccountBulkImportedPayload imports an account tree at once. Accounts must be
// in topological order (parents first).
type AccountBulkImportedPayload struct {
	Accounts []AccountCreatePayload `json:"accounts"`
}

func (p AccountBulkImportedPayload) EventType() event_types.EventType {
	return event_types.EventAccountBulkImported.Enum()
}

func (p AccountBulkImportedPayload) Validate() error {
	if len(p.Accounts) == 0 {
		return joinErrors([]string{"accounts is required"})
	}
	seen := make(map[string]bool, len(p.Accounts))
	for _, a := range p.Accounts {
		if err := a.Validate(); err != nil {
			return err
		}
		if seen[a.AccountId] {
			return joinErrors([]string{"duplicate account_id: " + a.AccountId})
		}
		seen[a.AccountId] = true
	}
	return nil
}

// AccountCreatePayload is the payload for account.created
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

func (p AccountCreatePayload) EventType() event_types.EventType {
	return event_types.EventAccountCreated.Enum()
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

// AccountUpdatedPayload is the payload for account.updated. Version is the
// account version the client read (ADR-027 rule 5).
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

func (p AccountUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventAccountUpdated.Enum()
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

// LedgerAccountCreatePayload is the payload for ledger_account.created
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

func (p LedgerAccountCreatePayload) EventType() event_types.EventType {
	return event_types.EventLedgerAccountCreated.Enum()
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

// LedgerAccountUpdatedPayload is the payload for ledger_account.updated. The
// ledger account is named by its integer ledger_id (ISSUE-031); Version is the
// version the client read.
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

func (p LedgerAccountUpdatedPayload) EventType() event_types.EventType {
	return event_types.EventLedgerAccountUpdated.Enum()
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
