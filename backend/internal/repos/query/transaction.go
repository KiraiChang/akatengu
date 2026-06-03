package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type TransactionRepo interface {
	GetByID(ctx context.Context, id int64) (*projection.Transaction, error)
	GetByUUID(ctx context.Context, txnUUID string) (*projection.Transaction, error)
	GetTransactionPaged(ctx context.Context, req model.PaginationParams) ([]projection.Transaction, int64, error)
	GetEntries(ctx context.Context, id int64) ([]projection.Entry, error)
}

type sqlcdbTransactionRepository struct {
	q *sqlcdb.Queries
}

func (r *sqlcdbTransactionRepository) GetEntries(ctx context.Context, id int64) ([]projection.Entry, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetJournalEntries(ctx, sqlcdb.GetJournalEntriesParams{
		TxnID:      id,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.Entry, len(rows))
	for i, row := range rows {
		result[i] = projection.EntryFromGetJournalEntriesRow(row)
	}
	return result, nil
}

func (r *sqlcdbTransactionRepository) GetTransactionPaged(ctx context.Context, req model.PaginationParams) ([]projection.Transaction, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetTransactionPaged(ctx, sqlcdb.GetTransactionPagedParams{
		MerchantID: merchantID,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.Transaction, len(rows))
	for i, row := range rows {
		result[i] = projection.TransactionFromGetTransactionPagedRow(row)
	}
	return result, total, nil
}

func newTransactionRepo(q *sqlcdb.Queries) TransactionRepo {
	return &sqlcdbTransactionRepository{q: q}
}

func NewTransactionRepo(db *sqlx.DB) TransactionRepo {
	return &sqlcdbTransactionRepository{q: sqlcdb.New(db)}
}

func (r *sqlcdbTransactionRepository) GetByUUID(ctx context.Context, txnUUID string) (*projection.Transaction, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetTransactionByUUID(ctx, sqlcdb.GetTransactionByUUIDParams{
		TxnUuid:    txnUUID,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.TransactionPtrFromGetTransactionByUUIDRow(row), nil
}

func (r *sqlcdbTransactionRepository) GetByID(ctx context.Context, id int64) (*projection.Transaction, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetTransaction(ctx, sqlcdb.GetTransactionParams{
		TxnID:      id,
		MerchantID: merchantID,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.TransactionPtrFromGetTransactionRow(row), nil
}
