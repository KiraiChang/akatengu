package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"context"
)

type ConfigProjectionRepo interface {
	UpsertSysAccount(ctx context.Context, p sqlcdb.UpsertSysAccountParams) error
	UpsertLedgerAccountTypeConfig(ctx context.Context, p sqlcdb.UpsertLedgerAccountTypeConfigParams) error
	UpsertAssetTypeAccountConfig(ctx context.Context, p sqlcdb.UpsertAssetTypeAccountConfigParams) error
}

type sqlxConfigProjectionRepo struct {
	q *sqlcdb.Queries
}

func NewConfigProjectionRepo(q *sqlcdb.Queries) ConfigProjectionRepo {
	return &sqlxConfigProjectionRepo{q: q}
}

func (r *sqlxConfigProjectionRepo) UpsertSysAccount(ctx context.Context, p sqlcdb.UpsertSysAccountParams) error {
	return r.q.UpsertSysAccount(ctx, p)
}

func (r *sqlxConfigProjectionRepo) UpsertLedgerAccountTypeConfig(ctx context.Context, p sqlcdb.UpsertLedgerAccountTypeConfigParams) error {
	return r.q.UpsertLedgerAccountTypeConfig(ctx, p)
}

func (r *sqlxConfigProjectionRepo) UpsertAssetTypeAccountConfig(ctx context.Context, p sqlcdb.UpsertAssetTypeAccountConfigParams) error {
	return r.q.UpsertAssetTypeAccountConfig(ctx, p)
}
