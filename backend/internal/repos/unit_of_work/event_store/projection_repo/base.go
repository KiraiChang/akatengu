package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
)

type TxProjectionRepository struct {
	AccountRepo     AccountRepo
	TransactionRepo TransactionRepo
	InvestmentRepo  InvestmentRepo
	PeriodCloseRepo PeriodCloseRepo
	InstallmentRepo InstallmentRepo
}

func NewTxProjectionRepository(q *sqlcdb.Queries) *TxProjectionRepository {
	return &TxProjectionRepository{
		AccountRepo:     NewAccountRepo(q),
		TransactionRepo: NewTransactionRepo(q),
		InvestmentRepo:  NewInvestmentRepo(q),
		PeriodCloseRepo: NewPeriodCloseRepo(q),
		InstallmentRepo: NewInstallment(q),
	}
}