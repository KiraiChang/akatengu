package projection_repo

import (
	"github.com/jmoiron/sqlx"
)

type TxProjectionRepository struct {
	AccountRepo     AccountRepo
	TransactionRepo TransactionRepo
	InvestmentRepo  InvestmentRepo
	PeriodCloseRepo PeriodCloseRepo
}

func NewTxProjectionRepository(tx *sqlx.Tx) *TxProjectionRepository {
	return &TxProjectionRepository{
		AccountRepo:     NewAccountRepo(tx),
		TransactionRepo: NewTransactionRepo(tx),
		InvestmentRepo:  NewInvestmentRepo(tx),
		PeriodCloseRepo: NewPeriodCloseRepo(tx),
	}
}
