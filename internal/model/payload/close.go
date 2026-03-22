package payload

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"

	"github.com/shopspring/decimal"
)

type PeriodCloseStartedPayload struct {
	PeriodType  enums.PeriodType `json:"period_type"` // 'monthly' | 'annual'
	PeriodStart string           `json:"period_start"`
	PeriodEnd   string           `json:"period_end"`
}

type PeriodClosedPayload struct {
	ClosingId   int64            `json:"closing_id"`
	PeriodType  enums.PeriodType `json:"period_type"`
	PeriodStart string           `json:"period_start"`
	PeriodEnd   string           `json:"period_end"`
	ClosedAt    string           `json:"closed_at"`
}

type PeriodReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}

type AnnualClosingEntryCreatedPayload struct {
	Year          int                `json:"year"`
	TransactionId int64              `json:"txn_id"`
	NetIncome     decimal.Decimal    `json:"net_income"`
	Entries       []projection.Entry `json:"entries"`
}

type AnnualOpeningEntryCreatedPayload struct {
	Year          int                `json:"year"`
	TransactionId int64              `json:"txn_id"`
	NetIncome     decimal.Decimal    `json:"net_income"`
	Entries       []projection.Entry `json:"entries"`
}
