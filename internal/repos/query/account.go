package query

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type AccountRepo interface {
	GetAccount(ctx context.Context, id int64) (*projection.Account, error)
	GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error)
}

type sqlxAccountRepo struct {
	db *sqlx.DB
}

func (s sqlxAccountRepo) GetAccount(ctx context.Context, id int64) (*projection.Account, error) {
	var result projection.Account
	if err := s.db.QueryRowxContext(ctx, `
        SELECT * FROM accounts
        WHERE account_id = ?`,
		id,
	).StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s sqlxAccountRepo) GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error) {
	var result projection.LedgerAccount
	if err := s.db.QueryRowxContext(ctx, `
        SELECT * FROM ledger_accounts
        WHERE ledger_id = ?`,
		id,
	).StructScan(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func NewAccountRepo(db *sqlx.DB) AccountRepo {
	return &sqlxAccountRepo{
		db: db,
	}
}
