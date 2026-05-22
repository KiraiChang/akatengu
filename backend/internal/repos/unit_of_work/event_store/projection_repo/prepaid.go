package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"context"

	"github.com/shopspring/decimal"
)

type sqlxPrepaidRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxPrepaidRepo) InsertPrepaid(ctx context.Context, p sqlcdb.InsertPrepaidParams) (int64, error) {
	row, err := r.q.InsertPrepaid(ctx, p)
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *sqlxPrepaidRepo) UpdatePrepaidTxn(ctx context.Context, id int64, merchantID int64, txnID int64) error {
	return r.q.UpdatePrepaidTxn(ctx, sqlcdb.UpdatePrepaidTxnParams{
		ID:         id,
		MerchantID: merchantID,
		TxnID:      &txnID,
	})
}

func (r *sqlxPrepaidRepo) UpdatePrepaidAmortization(ctx context.Context, id int64, merchantID int64, deltaAmount decimal.Decimal, status enums.PrepaidStatus, updatedBy *string) error {
	return r.q.UpdatePrepaidAmortization(ctx, sqlcdb.UpdatePrepaidAmortizationParams{
		ID:          id,
		MerchantID:  merchantID,
		DeltaAmount: deltaAmount,
		Status:      status,
		UpdatedBy:   updatedBy,
	})
}

func (r *sqlxPrepaidRepo) UpdatePrepaidDisposed(ctx context.Context, id int64, merchantID int64, updatedBy *string) error {
	return r.q.UpdatePrepaidDisposed(ctx, sqlcdb.UpdatePrepaidDisposedParams{
		ID:         id,
		MerchantID: merchantID,
		UpdatedBy:  updatedBy,
	})
}

func (r *sqlxPrepaidRepo) InsertPrepaidAmortization(ctx context.Context, p sqlcdb.InsertPrepaidAmortizationParams) error {
	return r.q.InsertPrepaidAmortization(ctx, p)
}

func NewPrepaidRepo(q *sqlcdb.Queries) PrepaidRepo {
	return &sqlxPrepaidRepo{q}
}
