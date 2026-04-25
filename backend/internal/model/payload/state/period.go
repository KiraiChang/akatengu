package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
)

type PeriodMonthClosedState struct {
	Next     *projection.PeriodClosing `json:"next"`
	Snapshot string                    `json:"snapshot"`
}
type PeriodAnnualClosedState struct {
	ClosedTxn payload.TransactionCreatedPayload `json:"closed_txn"`
	OpenedTxn payload.TransactionCreatedPayload `json:"opened_txn"`
	Next      *projection.PeriodClosing         `json:"next"`
	Snapshot  string                            `json:"snapshot"`
}

type PeriodAnnualReopenedState struct {
	ReverseClosedTxn payload.TransactionCreatedPayload `json:"reverse_closed_txn"`
	ReverseOpenedTxn payload.TransactionCreatedPayload `json:"opened_txn"`
}
