package projection

import "context"

// The Delete methods clear one merchant's rows of a read model, so a rebuild
// can apply the merchant's events again (BFS R3 D1). Each projector's Reset
// calls the ones of its own tables.

func (r *sqlxProjection) DeleteAccountClosure(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllAccountClosureByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteLedgerAccounts(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllLedgerAccountsByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteAccounts(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllAccountsByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteSysAccounts(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllSysAccountsByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteLedgerAccountTypeConfigs(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllLedgerAccountTypeConfigByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteAssetTypeAccountConfigs(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllAssetTypeAccountConfigByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeletePrepaidCategories(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllPrepaidCategoriesByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteFixedAssetCategories(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllFixedAssetCategoriesByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeletePeriodClosings(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllPeriodClosingsByMerchant(ctx, merchantID)
}

func (r *sqlxProjection) DeleteEntryCFCategories(ctx context.Context, merchantID int64) error {
	return r.q.DeleteAllEntryCFCategoriesByMerchant(ctx, merchantID)
}
