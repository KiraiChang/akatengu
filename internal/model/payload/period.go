package payload

import "akatengu/internal/model/db/projection"

type PeriodMonthStartedPayload struct {
	PeriodStart string `json:"period_start"`
}

type PeriodMonthClosedPayload struct {
	ClosingId int64                     `json:"closing_id"`
	ClosedAt  string                    `json:"closed_at"`
	Next      *projection.PeriodClosing `json:"next"`
	Snapshot  string                    `json:"snapshot"`
}

type PeriodMonthReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}

type PeriodAnnualStartedPayload struct {
	Year int `json:"year"`
}

type PeriodAnnualClosedPayload struct {
	ClosingId         int64                     `json:"closing_id"`
	ClosedTransaction TransactionCreatedPayload `json:"closed_txn"`
	OpenedTransaction TransactionCreatedPayload `json:"opened_txn"`
	Next              *projection.PeriodClosing `json:"next"`
	Snapshot          string                    `json:"snapshot"`
	ClosedAt          string                    `json:"closed_at"`
}

type PeriodAnnualReopenedPayload struct {
	ClosingId        int64                     `json:"closing_id"`
	ReverseClosedTxn TransactionCreatedPayload `json:"reverse_closed_txn"`
	ReverseOpenedTxn TransactionCreatedPayload `json:"opened_txn"`
	Reason           string                    `json:"reason"`
	ReopenedAt       string                    `json:"reopened_at"`
}
