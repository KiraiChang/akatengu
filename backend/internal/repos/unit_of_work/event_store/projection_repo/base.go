package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
)

type TxProjectionRepository struct {
	AccountRepo                      AccountRepo
	TransactionRepo                  TransactionRepo
	InvestmentRepo                   InvestmentRepo
	PeriodCloseRepo                  PeriodCloseRepo
	InstallmentRepo                  InstallmentRepo
	AccountBalanceSnapshotRepo       AccountBalanceSnapshotRepo
	LedgerAccountBalanceSnapshotRepo LedgerAccountBalanceSnapshotRepo
	AccountRunningBalanceRepo        AccountRunningBalanceRepo
	LedgerRunningBalanceRepo         LedgerRunningBalanceRepo
}

func NewTxProjectionRepository(q *sqlcdb.Queries) *TxProjectionRepository {
	return &TxProjectionRepository{
		AccountRepo:                      NewAccountRepo(q),
		TransactionRepo:                  NewTransactionRepo(q),
		InvestmentRepo:                   NewInvestmentRepo(q),
		PeriodCloseRepo:                  NewPeriodCloseRepo(q),
		InstallmentRepo:                  NewInstallment(q),
		AccountBalanceSnapshotRepo:       NewAccountBalanceSnapshotRepo(q),
		LedgerAccountBalanceSnapshotRepo: NewLedgerAccountBalanceSnapshotRepo(q),
		AccountRunningBalanceRepo:        NewAccountRunningBalanceRepo(q),
		LedgerRunningBalanceRepo:         NewLedgerRunningBalanceRepo(q),
	}
}
