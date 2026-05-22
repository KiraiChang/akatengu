package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"

	"github.com/shopspring/decimal"
)

type sqlxFixedAssetRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxFixedAssetRepo) InsertFixedAsset(ctx context.Context, p sqlcdb.InsertFixedAssetParams) (int64, error) {
	row, err := r.q.InsertFixedAsset(ctx, p)
	if err != nil {
		return 0, err
	}
	return row.ID, nil
}

func (r *sqlxFixedAssetRepo) UpdateFixedAssetTxn(ctx context.Context, id int64, merchantID int64, txnID int64) error {
	return r.q.UpdateFixedAssetTxn(ctx, sqlcdb.UpdateFixedAssetTxnParams{
		ID:         id,
		MerchantID: merchantID,
		TxnID:      &txnID,
	})
}

func (r *sqlxFixedAssetRepo) UpdateFixedAssetDepreciation(ctx context.Context, id int64, merchantID int64, deltaAmount decimal.Decimal, updatedBy *string) error {
	return r.q.UpdateFixedAssetDepreciation(ctx, sqlcdb.UpdateFixedAssetDepreciationParams{
		ID:          id,
		MerchantID:  merchantID,
		DeltaAmount: deltaAmount,
		UpdatedBy:   updatedBy,
	})
}

func (r *sqlxFixedAssetRepo) UpdateFixedAssetDisposed(ctx context.Context, id int64, merchantID int64, disposalDate string, updatedBy *string) error {
	return r.q.UpdateFixedAssetDisposed(ctx, sqlcdb.UpdateFixedAssetDisposedParams{
		ID:           id,
		MerchantID:   merchantID,
		DisposalDate: &disposalDate,
		UpdatedBy:    updatedBy,
	})
}

func (r *sqlxFixedAssetRepo) InsertFixedAssetDepreciation(ctx context.Context, p sqlcdb.InsertFixedAssetDepreciationParams) error {
	return r.q.InsertFixedAssetDepreciation(ctx, p)
}

func NewFixedAssetRepo(q *sqlcdb.Queries) FixedAssetRepo {
	return &sqlxFixedAssetRepo{q}
}
