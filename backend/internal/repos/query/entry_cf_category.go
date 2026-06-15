package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type EntryCFCategoryRepo interface {
	ListTransactionCFReview(ctx context.Context, dateFrom, dateTo string) ([]projection.TransactionCFReview, error)
}

type sqlcdbEntryCFCategoryRepo struct {
	q *sqlcdb.Queries
}

func newEntryCFCategoryRepo(q *sqlcdb.Queries) EntryCFCategoryRepo {
	return &sqlcdbEntryCFCategoryRepo{q: q}
}

func (r *sqlcdbEntryCFCategoryRepo) ListTransactionCFReview(ctx context.Context, dateFrom, dateTo string) ([]projection.TransactionCFReview, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListTransactionCFReview(ctx, sqlcdb.ListTransactionCFReviewParams{
		MerchantID: merchantID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
	if err != nil {
		return nil, err
	}
	results := make([]projection.TransactionCFReview, len(rows))
	for i, row := range rows {
		results[i] = projection.TransactionCFReviewFromListTransactionCFReviewRow(row)
	}
	return results, nil
}
