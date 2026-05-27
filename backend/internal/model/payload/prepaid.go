package payload

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"fmt"

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

// AmortizationAmount calculates this period's amortization amount.
// Last period gets the remainder to avoid decimal drift.
func AmortizationAmount(total decimal.Decimal, periods int64, amortizedPeriods int64, alreadyAmortized decimal.Decimal) decimal.Decimal {
	remaining := periods - amortizedPeriods
	if remaining <= 1 {
		return total.Sub(alreadyAmortized)
	}
	return total.Div(decimal.NewFromInt(periods)).Truncate(6)
}

// BuildPrepaidCreatedTransaction assembles the journal entry payload for EventPrepaidCreated.
// Called by the pipeline factory; the result is stored in PrepaidCreatedState.Transaction.
func BuildPrepaidCreatedTransaction(p PrepaidCreatedPayload, ledger *projection.LedgerAccount) TransactionCreatedPayload {
	cfOperating := enums.CashFlowCategoryOperating.Enum()
	ledgerId := ledger.LedgerId
	return TransactionCreatedPayload{
		TransactionDate: p.StartDate,
		Description:     fmt.Sprintf("預付費用 %s", p.Name),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: p.AccountID, Debit: p.TotalAmount, Credit: decimal.Zero, CashFlowCategory: &cfOperating},
			{AccountId: ledger.AccountId, LedgerId: &ledgerId, Debit: decimal.Zero, Credit: p.TotalAmount},
		},
	}
}

// BuildPrepaidAmortizedTransaction assembles the journal entry payload for EventPrepaidAmortized.
// Called by the pipeline factory; the result is stored in PrepaidAmortizedState.Transaction.
func BuildPrepaidAmortizedTransaction(p PrepaidAmortizedPayload, prepaid *projection.Prepaid) TransactionCreatedPayload {
	cfOperating := enums.CashFlowCategoryOperating.Enum()
	amortAmount := AmortizationAmount(prepaid.TotalAmount, prepaid.Periods, prepaid.AmortizedPeriods, prepaid.AmortizedAmount)
	return TransactionCreatedPayload{
		TransactionDate: p.PeriodDate + "-01",
		Description:     fmt.Sprintf("預付費用攤提 %s %s", prepaid.Name, p.PeriodDate),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: prepaid.ExpenseAccountID, Debit: amortAmount, Credit: decimal.Zero},
			{AccountId: prepaid.AccountID, Debit: decimal.Zero, Credit: amortAmount, CashFlowCategory: &cfOperating},
		},
	}
}

// BuildPrepaidDisposedTransaction assembles the journal entry payload for EventPrepaidDisposed.
// Called by the pipeline factory; the result is stored in PrepaidDisposedState.Transaction.
func BuildPrepaidDisposedTransaction(p PrepaidDisposedPayload, prepaid *projection.Prepaid) TransactionCreatedPayload {
	cfOperating := enums.CashFlowCategoryOperating.Enum()
	remaining := prepaid.TotalAmount.Sub(prepaid.AmortizedAmount)
	return TransactionCreatedPayload{
		TransactionDate: p.DisposalDate,
		Description:     fmt.Sprintf("預付費用提前終止 %s", prepaid.Name),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: prepaid.ExpenseAccountID, Debit: remaining, Credit: decimal.Zero},
			{AccountId: prepaid.AccountID, Debit: decimal.Zero, Credit: remaining, CashFlowCategory: &cfOperating},
		},
	}
}
