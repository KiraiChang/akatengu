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
	PrepaidRepo                      PrepaidRepo
	FixedAssetRepo                   FixedAssetRepo
	FixedAssetCategoryRepo           FixedAssetCategoryRepo
	PrepaidCategoryRepo              PrepaidCategoryRepo
	ConfigProjectionRepo             ConfigProjectionRepo
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
		PrepaidRepo:                      NewPrepaidRepo(q),
		FixedAssetRepo:                   NewFixedAssetRepo(q),
		FixedAssetCategoryRepo:           NewFixedAssetCategoryRepo(q),
		PrepaidCategoryRepo:              NewPrepaidCategoryRepo(q),
		ConfigProjectionRepo:             NewConfigProjectionRepo(q),
		AccountBalanceSnapshotRepo:       NewAccountBalanceSnapshotRepo(q),
		LedgerAccountBalanceSnapshotRepo: NewLedgerAccountBalanceSnapshotRepo(q),
		AccountRunningBalanceRepo:        NewAccountRunningBalanceRepo(q),
		LedgerRunningBalanceRepo:         NewLedgerRunningBalanceRepo(q),
	}
}
