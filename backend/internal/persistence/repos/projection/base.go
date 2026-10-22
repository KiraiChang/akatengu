package projection

import (
	"akatengu/internal/database/sqlcdb"
	dbprojection "akatengu/internal/model/db/projection"
	"context"
)

type Projection interface {
	UpsertEntryCFCategory(ctx context.Context, p sqlcdb.UpsertEntryCFCategoryParams) error

	CreateAccount(ctx context.Context, p dbprojection.Account) error
	UpdateAccount(ctx context.Context, p dbprojection.Account) error
	UpsertAccount(ctx context.Context, p dbprojection.Account) error
	CreateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error
	UpdateLedgerAccount(ctx context.Context, p dbprojection.LedgerAccount) error

	UpsertSysAccount(ctx context.Context, p sqlcdb.UpsertSysAccountParams) error
	UpsertLedgerAccountTypeConfig(ctx context.Context, p sqlcdb.UpsertLedgerAccountTypeConfigParams) error
	UpsertAssetTypeAccountConfig(ctx context.Context, p sqlcdb.UpsertAssetTypeAccountConfigParams) error

	InsertPrepaidCategory(ctx context.Context, p sqlcdb.InsertPrepaidCategoryParams) error
	UpdatePrepaidCategory(ctx context.Context, p sqlcdb.UpdatePrepaidCategoryParams) error
	SoftDeletePrepaidCategory(ctx context.Context, p sqlcdb.SoftDeletePrepaidCategoryParams) error
	InsertFixedAssetCategory(ctx context.Context, p sqlcdb.InsertFixedAssetCategoryParams) error
	UpdateFixedAssetCategory(ctx context.Context, p sqlcdb.UpdateFixedAssetCategoryParams) error
	SoftDeleteFixedAssetCategory(ctx context.Context, p sqlcdb.SoftDeleteFixedAssetCategoryParams) error

	InsertPeriodClose(ctx context.Context, p dbprojection.PeriodClosing) error
}

type sqlxProjection struct {
	q *sqlcdb.Queries
}

func NewProjection(q *sqlcdb.Queries) Projection {
	return &sqlxProjection{q: q}
}
