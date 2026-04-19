package payload

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// InstallmentCreatedPayload
// ─────────────────────────────────────────

type InstallmentCreatedPayload struct {
	Amount           decimal.Decimal    `json:"amount"`                // "10000"
	InstallmentCount int                `json:"installment_count"`     // 12
	StartDate        string             `json:"start_date"`            // "2026-05-01"，第一期到期日
	InterestType     enums.InterestType `json:"interest_type"`         // "interest_free" | "fixed_rate"
	AnnualRate       decimal.Decimal    `json:"annual_rate,omitempty"` // "12.5"，有息才填
	AccountId        string             `json:"account_id"`            // 買的東西歸屬科目，如「電腦設備」
	LedgerId         int64              `json:"ledger_id"`             // 哪張信用卡 / 哪個帳戶
	Memo             string             `json:"memo,omitempty"`
	Note             string             `json:"note,omitempty"`
}

func (p InstallmentCreatedPayload) Validate() error {
	var errs []string

	if p.Amount.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "amount must be greater than zero")
	}

	if p.InstallmentCount <= 0 {
		errs = append(errs, "installment_count is required")
	}

	if p.StartDate == "" {
		errs = append(errs, "start_date is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}

	if p.LedgerId <= 0 {
		errs = append(errs, "ledger_id is required")
	}

	if p.InterestType != enums.InterestTypeFree.Enum() ||
		p.InterestType != enums.InterestTypeFixedRate.Enum() {
		errs = append(errs, "interest_type is required")
	}

	return joinErrors(errs)
}

func (p InstallmentCreatedPayload) CreateInstallment() (*projection.Installment, error) {
	return &projection.Installment{
		LedgerId:        p.LedgerId,
		Description:     p.Memo,
		TotalAmount:     p.Amount,
		TotalPeriods:    p.InstallmentCount,
		PaidPeriods:     0,
		AmountPerPeriod: p.Amount.Div(decimal.NewFromInt(int64(p.InstallmentCount))),
		StartDate:       p.StartDate,
		InterestRate:    p.AnnualRate,
		InterestType:    p.InterestType,
		Status:          enums.InstallmentStatusActive.Enum(),
		Note:            p.Note,
	}, nil
}

func (p InstallmentCreatedPayload) CreatePayments() ([]*projection.InstallmentPayment, error) {
	switch p.InterestType.Val() {
	case enums.InterestTypeFree:
		return p.generateInterestFreeItems()
	case enums.InterestTypeFixedRate:
		return p.generateFixedRateItems()
	}
	return nil, fmt.Errorf("unknown interest type: %s", p.InterestType)
}

func (p InstallmentCreatedPayload) generateInterestFreeItems() ([]*projection.InstallmentPayment, error) {
	dateStart, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil, err
	}

	items := make([]*projection.InstallmentPayment, p.InstallmentCount)
	base := p.Amount.Div(decimal.NewFromInt(int64(p.InstallmentCount))).
		Truncate(6)
	// 最後一期補尾差
	remainder := p.Amount.Sub(base.Mul(decimal.NewFromInt(int64(p.InstallmentCount - 1))))

	for i := 0; i < p.InstallmentCount; i++ {
		amount := base
		if i == p.InstallmentCount-1 {
			amount = remainder
		}
		items[i] = &projection.InstallmentPayment{
			Period:   i + 1,
			DueDate:  dateStart.AddDate(0, i+1, 0).Format("2006-01-02"),
			Status:   enums.InstallmentPaymentStatusPending.Enum(),
			Amount:   amount,
			Interest: decimal.Zero,
		}
	}
	return items, nil
}

// generateFixedRateItems 平息法：每期利息 = 本金 × 月利率，本金均攤
func (p InstallmentCreatedPayload) generateFixedRateItems() ([]*projection.InstallmentPayment, error) {
	dateStart, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil, err
	}

	n := int64(p.InstallmentCount)
	monthlyRate := p.AnnualRate.Div(decimal.NewFromInt(12)).Div(decimal.NewFromInt(100))

	basePrincipal := p.Amount.Div(decimal.NewFromInt(n)).Truncate(6)
	principalRemainder := p.Amount.Sub(basePrincipal.Mul(decimal.NewFromInt(n - 1)))

	monthlyInterest := p.Amount.Mul(monthlyRate).Truncate(6)
	interestRemainder := p.Amount.Sub(monthlyInterest.Mul(decimal.NewFromInt(n - 1)))

	items := make([]*projection.InstallmentPayment, p.InstallmentCount)
	for i := 0; i < p.InstallmentCount; i++ {
		principal := basePrincipal
		interest := monthlyInterest
		if i == p.InstallmentCount-1 {
			principal = principalRemainder
			interest = interestRemainder
		}
		items[i] = &projection.InstallmentPayment{
			Period:   i + 1,
			DueDate:  dateStart.AddDate(0, i+1, 0).Format("2006-01-02"),
			Status:   enums.InstallmentPaymentStatusPending.Enum(),
			Amount:   principal,
			Interest: interest,
		}
	}
	return items, nil
}

// ─────────────────────────────────────────
// InstallmentPeriodPaidPayload
// ─────────────────────────────────────────

type InstallmentPeriodPaidPayload struct {
	InstallmentId int64  `json:"installment_id"` // 1
	Period        int    `json:"period"`         // 期數
	PaidDate      string `json:"paid_date"`      // "2026-05-01"，到期日
	PaidLedgerId  int64  `json:"paid_ledger_id"` // 由那裡扣款 現金/信用卡當期帳單
}

func (p InstallmentPeriodPaidPayload) Validate() error {
	var errs []string

	if p.InstallmentId == 0 {
		errs = append(errs, "installment_id is required")
	}

	if p.Period == 0 {
		errs = append(errs, "period is required")
	}

	if p.PaidDate == "" {
		errs = append(errs, "PaidDate is required")
	}

	if p.PaidLedgerId == 0 {
		errs = append(errs, "paid_ledger_id is required")
	}

	return joinErrors(errs)
}
