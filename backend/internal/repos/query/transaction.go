package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TransactionRepo interface {
	GetByID(ctx context.Context, id int64) (*projection.Transaction, error)
}

type sqlcdbTransactionRepository struct {
	q *sqlcdb.Queries
}

func newTransactionRepo(q *sqlcdb.Queries) TransactionRepo {
	return &sqlcdbTransactionRepository{q: q}
}

func NewTransactionRepo(db *sqlx.DB) TransactionRepo {
	return &sqlcdbTransactionRepository{q: sqlcdb.New(db)}
}

func (r *sqlcdbTransactionRepository) GetByID(ctx context.Context, id int64) (*projection.Transaction, error) {
	row, err := r.q.GetTransaction(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.TransactionPtrFromTransaction(row), nil
}
