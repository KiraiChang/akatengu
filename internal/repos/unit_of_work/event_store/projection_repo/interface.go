package projection_repo

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"context"
)

type TransactionRepo interface {
	InsertTxn(ctx context.Context, p projection.Transaction) (int64, error)
	UpdateTxnStatus(ctx context.Context, txn_id int64, status enums.TransactionStatus, version int64) error
	SysUpdateTxnStatus(ctx context.Context, txn_id int64, refTxnId *int64, status enums.TransactionStatus) error
	UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error
}

type InstallmentRepo interface {
	//// installment
	//UpsertInstallment(ctx context.Context, p ProjInstallment) error
	//UpsertInstallmentPayment(ctx context.Context, p ProjInstallmentPayment) error
	//
	//// reconciliation
	//UpsertReconciliation(ctx context.Context, p ProjReconciliation) error
	//UpsertReconciliationAdjustment(ctx context.Context, p ProjReconciliationAdjustment) error
}

type AccountRepo interface {
	// account
	CreateAccount(ctx context.Context, p projection.Account) error
	UpdateAccount(ctx context.Context, p projection.Account) error
	CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error
	UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error
}

//// 重建用
//TruncateAll(ctx context.Context) error
//TruncateByAggregate(ctx context.Context, aggregateType enums.AggregateType) error

type InvestmentRepo interface {
	// investment
	CreateInvestment(ctx context.Context, p projection.Investment) error
	UpdateInvestment(ctx context.Context, p projection.Investment) error
	UpsertExchangeRate(ctx context.Context, rate projection.ExchangeRate) error
	UpsertPosition(ctx context.Context, position projection.InvestmentPosition) error
	InsertLot(ctx context.Context, lot projection.InvestmentLot) (int64, error)
	InsertLotDisposals(ctx context.Context, disposals projection.InvestmentLotDisposals) error
	InsertMovement(ctx context.Context, movement projection.InvestmentMovement) (int64, error)
	UpdateMovement(ctx context.Context, id int64, txnId int64) error
	UpdateLot(ctx context.Context, id int64, txnId int64) error
}

type PeriodCloseRepo interface {
	InsertPeriodClose(ctx context.Context, c projection.PeriodClosing) (*int64, error)
	UpdatePeriodCloseSnapshotByPeriod(ctx context.Context, id int64, snapshot string, closedAt *string) error
	ReopenPeriodClose(ctx context.Context, id int64, reason string, at string) error
	UpdatePeriodCloseTxnId(ctx context.Context, id int64, closingTxnId *int64, openingTxnId *int64) error
}
