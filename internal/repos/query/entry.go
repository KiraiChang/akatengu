package query

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type EntryRepo interface {
	GetEntries(ctx context.Context, txnID int64) ([]projection.Entry, error)
}

type sqlxEntryRepo struct {
	db *sqlx.DB
}

func NewEntryRepo(db *sqlx.DB) EntryRepo {
	return &sqlxEntryRepo{db: db}
}

func (r *sqlxEntryRepo) GetEntries(ctx context.Context, txnID int64) ([]projection.Entry, error) {
	var entries []projection.Entry
	err := r.db.SelectContext(ctx, &entries, `
        SELECT account_id, ledger_id, debit, credit
        FROM journal_entries
        WHERE txn_id = ?`,
		txnID,
	)
	return entries, err
}
