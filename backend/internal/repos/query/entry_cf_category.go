package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/pkg/ctxkey"
	"context"
)

type EntryCFCategoryRepo interface {
	ListTransactionCFReview(ctx context.Context, dateFrom, dateTo string) ([]sqlcdb.ListTransactionCFReviewRow, error)
}

type sqlcdbEntryCFCategoryRepo struct {
	q *sqlcdb.Queries
}

func newEntryCFCategoryRepo(q *sqlcdb.Queries) EntryCFCategoryRepo {
	return &sqlcdbEntryCFCategoryRepo{q: q}
}

func (r *sqlcdbEntryCFCategoryRepo) ListTransactionCFReview(ctx context.Context, dateFrom, dateTo string) ([]sqlcdb.ListTransactionCFReviewRow, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.ListTransactionCFReview(ctx, sqlcdb.ListTransactionCFReviewParams{
		MerchantID: merchantID,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	})
}
