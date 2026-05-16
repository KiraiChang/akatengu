package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"

	"github.com/shopspring/decimal"
)

// AccountRunningBalanceRepo implementation

type sqlcdbAccountRunningBalanceRepo struct {
	q *sqlcdb.Queries
}

func NewAccountRunningBalanceRepo(q *sqlcdb.Queries) AccountRunningBalanceRepo {
	return &sqlcdbAccountRunningBalanceRepo{q: q}
}

func (r *sqlcdbAccountRunningBalanceRepo) Upsert(ctx context.Context, accountId string, merchantID int64, debit, credit decimal.Decimal) error {
	return r.q.UpsertAccountRunningBalance(ctx, sqlcdb.UpsertAccountRunningBalanceParams{
		AccountID:   accountId,
		MerchantID:  merchantID,
		DebitTotal:  debit,
		CreditTotal: credit,
	})
}

func (r *sqlcdbAccountRunningBalanceRepo) GetAncestorIds(ctx context.Context, accountId string, merchantID int64) ([]string, error) {
	return r.q.GetAncestorAccountIds(ctx, sqlcdb.GetAncestorAccountIdsParams{
		DescendantID: accountId,
		MerchantID:   merchantID,
	})
}

// LedgerRunningBalanceRepo implementation

type sqlcdbLedgerRunningBalanceRepo struct {
	q *sqlcdb.Queries
}

func NewLedgerRunningBalanceRepo(q *sqlcdb.Queries) LedgerRunningBalanceRepo {
	return &sqlcdbLedgerRunningBalanceRepo{q: q}
}

func (r *sqlcdbLedgerRunningBalanceRepo) Upsert(ctx context.Context, ledgerId int64, merchantID int64, debit, credit decimal.Decimal) error {
	return r.q.UpsertLedgerRunningBalance(ctx, sqlcdb.UpsertLedgerRunningBalanceParams{
		LedgerID:    ledgerId,
		MerchantID:  merchantID,
		DebitTotal:  debit,
		CreditTotal: credit,
	})
}
