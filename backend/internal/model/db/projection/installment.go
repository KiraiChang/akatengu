package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=Installment
//dbmap:sqlcdb=InsertInstallmentParams
//dbmap:sqlcdb=GetInstallmentRow
//dbmap:sqlcdb=GetInstallmentPagedRow
type Installment struct {
	MerchantID      int64                   `db:"merchant_id" json:"merchant_id"`
	InstallmentId   int64                   `db:"installment_id" json:"installment_id"`
	InstallmentUuid string                  `db:"installment_uuid" json:"installment_uuid"`
	TransactionId   *int64                  `db:"txn_id" json:"transaction_id"`
	LedgerId        int64                   `db:"ledger_id" json:"ledger_id"`
	Description     string                  `db:"description" json:"description"`
	TotalAmount     decimal.Decimal         `db:"total_amount" json:"total_amount"`
	TotalPeriods    int64                   `db:"total_periods" json:"total_periods"`
	PaidPeriods     int64                   `db:"paid_periods" json:"paid_periods"`
	AmountPerPeriod decimal.Decimal         `db:"amount_per_period" json:"amount_per_period"`
	StartDate       string                  `db:"start_date" json:"start_date"`
	EndDate         *string                 `db:"end_date" json:"end_date"`
	InterestRate    decimal.Decimal         `db:"interest_rate" json:"interest_rate"`
	InterestType    enums.InterestType      `db:"interest_type" json:"interest_type"`
	Status          enums.InstallmentStatus `db:"status" json:"status"`
	Note            string                  `db:"note" json:"note"`
	UpdatedBy       *string                 `db:"updated_by" json:"updated_by"`
	UpdatedAt       *string                 `db:"updated_at" json:"updated_at"`
}

//dbmap:sqlcdb=InstallmentPayment
//dbmap:sqlcdb=InsertInstallmentPaymentParams
//dbmap:sqlcdb=GetPaymentPagedRow
//dbmap:sqlcdb=GetInstallmentPaymentRow
type InstallmentPayment struct {
	MerchantID      int64                          `db:"merchant_id" json:"merchant_id"`
	PaymentId       int64                          `db:"payment_id" json:"payment_id"`
	PaymentUuid     string                         `db:"payment_uuid" json:"payment_uuid"`
	InstallmentId   int64                          `db:"installment_id" json:"installment_id"`
	InstallmentUuid string                         `db:"installment_uuid" json:"installment_uuid"`
	TransactionId   *int64                         `db:"txn_id" json:"transaction_id"`
	Period          int64                          `db:"period_no" json:"period"`
	Amount          decimal.Decimal                `db:"amount" json:"amount"`
	Interest        decimal.Decimal                `db:"interest" json:"interest"`
	DueDate         string                         `db:"due_date" json:"due_date"`
	PaidDate        *string                        `db:"paid_date" json:"paid_date"`
	Status          enums.InstallmentPaymentStatus `db:"status" json:"status"`
	UpdatedBy       *string                        `db:"updated_by" json:"updated_by"`
	UpdatedAt       *string                        `db:"updated_at" json:"updated_at"`
}
