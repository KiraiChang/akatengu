package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type Installment struct {
	InstallmentId   int64                   `db:"installment_id"`
	TransactionId   *int64                  `db:"txn_id"`
	LedgerId        int64                   `db:"ledger_id"`
	Description     string                  `db:"description"`
	TotalAmount     decimal.Decimal         `db:"total_amount"`
	TotalPeriods    int                     `db:"total_periods"`
	PaidPeriods     int                     `db:"paid_periods"`
	AmountPerPeriod decimal.Decimal         `db:"amount_per_period"`
	StartDate       string                  `db:"start_date"`
	EndDate         *string                 `db:"end_date"`
	InterestRate    decimal.Decimal         `db:"interest_rate"`
	InterestType    enums.InterestType      `db:"interest_type"`
	Status          enums.InstallmentStatus `db:"status"`
	Note            string                  `db:"note"`
}

type InstallmentPayment struct {
	PaymentId     int64                          `db:"payment_id"`
	InstallmentId int64                          `db:"installment_id"`
	TransactionId *int64                         `db:"txn_id"`
	Period        int                            `db:"period_no"`
	Amount        decimal.Decimal                `db:"amount"`
	Interest      decimal.Decimal                `db:"interest"`
	DueDate       string                         `db:"due_date"`
	PaidDate      *string                        `db:"paid_date"`
	Status        enums.InstallmentPaymentStatus `db:"status"`
}
