package projection

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

// The config methods run the same sqlc queries as the legacy
// projection_repo/config.go.

func (r *sqlxProjection) UpsertSysAccount(ctx context.Context, p sqlcdb.UpsertSysAccountParams) error {
	return r.q.UpsertSysAccount(ctx, p)
}

func (r *sqlxProjection) UpsertLedgerAccountTypeConfig(ctx context.Context, p sqlcdb.UpsertLedgerAccountTypeConfigParams) error {
	return r.q.UpsertLedgerAccountTypeConfig(ctx, p)
}

func (r *sqlxProjection) UpsertAssetTypeAccountConfig(ctx context.Context, p sqlcdb.UpsertAssetTypeAccountConfigParams) error {
	return r.q.UpsertAssetTypeAccountConfig(ctx, p)
}
