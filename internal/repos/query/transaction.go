package query

import (
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TransactionRepo interface {
	// 查詢

	GetByID(ctx context.Context, id int64) (*projection.Transaction, error)
}

// ─────────────────────────────────────────
// sqlx 實作
// ─────────────────────────────────────────

type sqlxTransactionRepository struct{ db *sqlx.DB }

func NewTransactionRepo(db *sqlx.DB) TransactionRepo {
	return &sqlxTransactionRepository{db: db}
}

func (r *sqlxTransactionRepository) GetByID(ctx context.Context, id int64) (*projection.Transaction, error) {
	var inv projection.Transaction
	err := r.db.QueryRowxContext(ctx, `SELECT * FROM transaction WHERE txn_id = ?`, id).
		StructScan(&inv)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}
