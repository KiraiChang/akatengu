package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type EntryRepo interface {
	GetEntries(ctx context.Context, txnID int64) ([]projection.Entry, error)
}

type sqlcdbEntryRepo struct {
	q *sqlcdb.Queries
}

func newEntryRepo(q *sqlcdb.Queries) EntryRepo {
	return &sqlcdbEntryRepo{q: q}
}

func NewEntryRepo(db *sqlx.DB) EntryRepo {
	return &sqlcdbEntryRepo{q: sqlcdb.New(db)}
}

func (r *sqlcdbEntryRepo) GetEntries(ctx context.Context, txnID int64) ([]projection.Entry, error) {
	rows, err := r.q.GetJournalEntries(ctx, txnID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Entry, len(rows))
	for i, row := range rows {
		result[i] = projection.EntryFromJournalEntry(row)
	}
	return result, nil
}
