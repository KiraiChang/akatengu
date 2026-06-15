package payload

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"fmt"

	"github.com/shopspring/decimal"
)

// PrepaidCreatedPayload 建立預付費用
type PrepaidCreatedPayload struct {
	CategoryUUID string          `json:"category_uuid"` // 預付費用類別 UUID
	LedgerUUID   string          `json:"ledger_uuid"`   // 付款帳戶
	Name         string          `json:"name"`          // 描述，如「保險費 2026-05 ～ 2027-04」
	TotalAmount  decimal.Decimal `json:"total_amount"`  // 預付總金額
	Periods      int64           `json:"periods"`       // 攤提期數（月）
	StartDate    string          `json:"start_date"`    // 開始攤提月份 YYYY-MM-DD
	Memo         string          `json:"memo,omitempty"`
	Note         string          `json:"note,omitempty"`
}

func (p PrepaidCreatedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	if p.LedgerUUID == "" {
		errs = append(errs, "ledger_uuid is required")
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

// PrepaidCreatedWithInstallmentPayload 以信用卡分期支付預付費用
type PrepaidCreatedWithInstallmentPayload struct {
	CategoryUUID string                  `json:"category_uuid"` // 預付費用類別 UUID
	Name         string                  `json:"name"`
	TotalAmount  decimal.Decimal         `json:"total_amount"`
	Periods      int64                   `json:"periods"`
	StartDate    string                  `json:"start_date"`
	Memo         string                  `json:"memo,omitempty"`
	Note         string                  `json:"note,omitempty"`
	Installment  InstallmentTermsPayload `json:"installment"`
}

func (p PrepaidCreatedWithInstallmentPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
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
	if err := p.Installment.Validate(); err != nil {
		errs = append(errs, err.Error())
	}
	return joinErrors(errs)
}

// ToInstallmentPayload 組合完整的 InstallmentCreatedPayload，以預付科目與總金額填入 AccountId / Amount。
// accountID 從類別設定取得，由 Pipeline 傳入。
func (p PrepaidCreatedWithInstallmentPayload) ToInstallmentPayload(accountID string) InstallmentCreatedPayload {
	return InstallmentCreatedPayload{
		Amount:           p.TotalAmount,
		InstallmentCount: p.Installment.InstallmentCount,
		StartDate:        p.Installment.StartDate,
		InterestType:     p.Installment.InterestType,
		AnnualRate:       p.Installment.AnnualRate,
		AccountId:        accountID,
		LedgerUuid:       p.Installment.LedgerUuid,
		Memo:             p.Installment.Memo,
		Note:             p.Installment.Note,
	}
}

// BuildPrepaidCreatedWithInstallmentTransaction 組合預付費用分期支付的會計分錄。
func BuildPrepaidCreatedWithInstallmentTransaction(p PrepaidCreatedWithInstallmentPayload, category *projection.PrepaidCategory, ledger *projection.LedgerAccount, sysAccountAssetPrepaidInterest string, payments []*projection.InstallmentPayment) (TransactionCreatedPayload, error) {
	ip := p.ToInstallmentPayload(category.AccountID)
	var entries []TransactionEntryPayload
	switch ip.InterestType.Val() {
	case enums.InterestTypeFree:
		entries = []TransactionEntryPayload{
			{AccountId: category.AccountID, Debit: p.TotalAmount, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledger.LedgerId, Debit: decimal.Zero, Credit: p.TotalAmount},
		}
	case enums.InterestTypeFixedRate:
		interest := decimal.Zero
		for _, pmt := range payments {
			interest = interest.Add(pmt.Interest)
		}
		entries = []TransactionEntryPayload{
			{AccountId: category.AccountID, Debit: p.TotalAmount, Credit: decimal.Zero},
			{AccountId: sysAccountAssetPrepaidInterest, Debit: interest, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledger.LedgerId, Debit: decimal.Zero, Credit: p.TotalAmount.Add(interest)},
		}
	default:
		return TransactionCreatedPayload{}, fmt.Errorf("invalid interest type: %s", ip.InterestType.Val())
	}
	return TransactionCreatedPayload{
		TransactionDate: p.StartDate,
		Description:     fmt.Sprintf("預付費用分期支付 %s", p.Name),
		Currency:        "TWD",
		Entries:         entries,
	}, nil
}

// PrepaidAmortizedPayload 執行一期攤提
type PrepaidAmortizedPayload struct {
	PrepaidUUID string `json:"prepaid_uuid"`
	PeriodDate  string `json:"period_date"` // 攤提月份 YYYY-MM
}

func (p PrepaidAmortizedPayload) Validate() error {
	var errs []string
	if p.PrepaidUUID == "" {
		errs = append(errs, "prepaid_uuid is required")
	}
	if p.PeriodDate == "" {
		errs = append(errs, "period_date is required")
	}
	return joinErrors(errs)
}

// PrepaidDisposedPayload 提前終止預付費用（一次認列剩餘金額）
type PrepaidDisposedPayload struct {
	PrepaidUUID  string `json:"prepaid_uuid"`
	DisposalDate string `json:"disposal_date"` // YYYY-MM-DD
	Memo         string `json:"memo,omitempty"`
}

func (p PrepaidDisposedPayload) Validate() error {
	var errs []string
	if p.PrepaidUUID == "" {
		errs = append(errs, "prepaid_uuid is required")
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
func BuildPrepaidCreatedTransaction(p PrepaidCreatedPayload, category *projection.PrepaidCategory, ledger *projection.LedgerAccount) TransactionCreatedPayload {
	ledgerId := ledger.LedgerId
	return TransactionCreatedPayload{
		TransactionDate: p.StartDate,
		Description:     fmt.Sprintf("預付費用 %s", p.Name),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: category.AccountID, Debit: p.TotalAmount, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledgerId, Debit: decimal.Zero, Credit: p.TotalAmount},
		},
	}
}

// BuildPrepaidAmortizedTransaction assembles the journal entry payload for EventPrepaidAmortized.
// Called by the pipeline factory; the result is stored in PrepaidAmortizedState.Transaction.
func BuildPrepaidAmortizedTransaction(p PrepaidAmortizedPayload, prepaid *projection.Prepaid) TransactionCreatedPayload {
	amortAmount := AmortizationAmount(prepaid.TotalAmount, prepaid.Periods, prepaid.AmortizedPeriods, prepaid.AmortizedAmount)
	return TransactionCreatedPayload{
		TransactionDate: p.PeriodDate + "-01",
		Description:     fmt.Sprintf("預付費用攤提 %s %s", prepaid.Name, p.PeriodDate),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: prepaid.ExpenseAccountID, Debit: amortAmount, Credit: decimal.Zero},
			{AccountId: prepaid.AccountID, Debit: decimal.Zero, Credit: amortAmount},
		},
	}
}

// BuildPrepaidDisposedTransaction assembles the journal entry payload for EventPrepaidDisposed.
// Called by the pipeline factory; the result is stored in PrepaidDisposedState.Transaction.
func BuildPrepaidDisposedTransaction(p PrepaidDisposedPayload, prepaid *projection.Prepaid) TransactionCreatedPayload {
	remaining := prepaid.TotalAmount.Sub(prepaid.AmortizedAmount)
	return TransactionCreatedPayload{
		TransactionDate: p.DisposalDate,
		Description:     fmt.Sprintf("預付費用提前終止 %s", prepaid.Name),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: prepaid.ExpenseAccountID, Debit: remaining, Credit: decimal.Zero},
			{AccountId: prepaid.AccountID, Debit: decimal.Zero, Credit: remaining},
		},
	}
}
