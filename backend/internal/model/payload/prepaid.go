package payload

import (
	"github.com/shopspring/decimal"
)

// PrepaidCreatedPayload 建立預付費用
type PrepaidCreatedPayload struct {
	AccountID        string          `json:"account_id"`         // 預付科目，如 1104-01
	ExpenseAccountID string          `json:"expense_account_id"` // 費用科目，如 5101-01
	LedgerID         int64           `json:"ledger_id"`          // 付款帳戶
	Name             string          `json:"name"`               // 描述，如「保險費 2026-05 ～ 2027-04」
	TotalAmount      decimal.Decimal `json:"total_amount"`       // 預付總金額
	Periods          int64           `json:"periods"`            // 攤提期數（月）
	StartDate        string          `json:"start_date"`         // 開始攤提月份 YYYY-MM-DD
	Memo             string          `json:"memo,omitempty"`
	Note             string          `json:"note,omitempty"`
}

func (p PrepaidCreatedPayload) Validate() error {
	var errs []string
	if p.AccountID == "" {
		errs = append(errs, "account_id is required")
	}
	if p.ExpenseAccountID == "" {
		errs = append(errs, "expense_account_id is required")
	}
	if p.LedgerID <= 0 {
		errs = append(errs, "ledger_id is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.TotalAmount.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "total_amount must be greater than zero")
	}
	if p.Periods <= 0 {
		errs = append(errs, "periods must be greater than zero")
	}
	if p.StartDate == "" {
		errs = append(errs, "start_date is required")
	}
	return joinErrors(errs)
}

// PrepaidAmortizedPayload 執行一期攤提
type PrepaidAmortizedPayload struct {
	PrepaidID  int64  `json:"prepaid_id"`
	PeriodDate string `json:"period_date"` // 攤提月份 YYYY-MM
}

func (p PrepaidAmortizedPayload) Validate() error {
	var errs []string
	if p.PrepaidID <= 0 {
		errs = append(errs, "prepaid_id is required")
	}
	if p.PeriodDate == "" {
		errs = append(errs, "period_date is required")
	}
	return joinErrors(errs)
}

// PrepaidDisposedPayload 提前終止預付費用（一次認列剩餘金額）
type PrepaidDisposedPayload struct {
	PrepaidID  int64  `json:"prepaid_id"`
	DisposalDate string `json:"disposal_date"` // YYYY-MM-DD
	Memo        string `json:"memo,omitempty"`
}

func (p PrepaidDisposedPayload) Validate() error {
	var errs []string
	if p.PrepaidID <= 0 {
		errs = append(errs, "prepaid_id is required")
	}
	if p.DisposalDate == "" {
		errs = append(errs, "disposal_date is required")
	}
	return joinErrors(errs)
}
