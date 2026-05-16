package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
)

type RunningBalanceRepo interface {
	GetAllLedger(ctx context.Context) ([]projection.LedgerAccountBalance, error)
	GetAllAccount(ctx context.Context) ([]projection.AccountBalance, error)
}

type sqlcdbRunningBalanceRepo struct {
	q *sqlcdb.Queries
}

func (s sqlcdbRunningBalanceRepo) GetAllLedger(ctx context.Context) ([]projection.LedgerAccountBalance, error) {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetAllLedgerRunningBalance(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccountBalance, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountBalanceFromGetAllLedgerRunningBalanceRow(row)
		if result[i].NormalBalance.Is(enums.BalanceDebit) {
			result[i].Balance = result[i].DebitTotal.Sub(result[i].CreditTotal)
		} else {
			result[i].Balance = result[i].CreditTotal.Sub(result[i].DebitTotal)
		}
	}
	return result, nil
}

func (s sqlcdbRunningBalanceRepo) GetAllAccount(ctx context.Context) ([]projection.AccountBalance, error) {
	merchantId, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetAllAccountRunningBalance(ctx, merchantId)
	if err != nil {
		return nil, err
	}
	result := make([]projection.AccountBalance, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountBalanceFromGetAllAccountRunningBalanceRow(row)
	}
	return result, nil
}

func NewRunningBalanceRepo(db *sqlx.DB) RunningBalanceRepo {
	return &sqlcdbRunningBalanceRepo{
		q: sqlcdb.New(db),
	}
}

func newRunningBalanceRepo(q *sqlcdb.Queries) RunningBalanceRepo {
	return &sqlcdbRunningBalanceRepo{
		q: q,
	}
}
