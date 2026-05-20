package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/pkg/dbmapconv"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
)

type ConfigRepo interface {
	GetLedgerAccountTypeConfigs(ctx context.Context) ([]sqlcdb.LedgerAccountTypeConfig, error)
	GetLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType) (*sqlcdb.LedgerAccountTypeConfig, error)
	UpsertLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType, accountID string, updatedBy *string) error

	GetAssetTypeAccountConfigs(ctx context.Context) ([]sqlcdb.AssetTypeAccountConfig, error)
	GetAssetTypeAccountConfig(ctx context.Context, at enums.AssetType) (*sqlcdb.AssetTypeAccountConfig, error)
	UpsertAssetTypeAccountConfig(ctx context.Context, params sqlcdb.UpsertAssetTypeAccountConfigParams) error

	GetAccountDescendants(ctx context.Context, ancestorID string) ([]projection.Account, error)
}

type sqlxConfigRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func newConfigRepo(q *sqlcdb.Queries, db *sqlx.DB) ConfigRepo {
	return &sqlxConfigRepo{q: q, db: db}
}

func NewConfigRepo(db *sqlx.DB) ConfigRepo {
	return &sqlxConfigRepo{q: sqlcdb.New(db), db: db}
}

func (r *sqlxConfigRepo) GetLedgerAccountTypeConfigs(ctx context.Context) ([]sqlcdb.LedgerAccountTypeConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.GetLedgerAccountTypeConfigs(ctx, merchantID)
}

func (r *sqlxConfigRepo) GetLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType) (*sqlcdb.LedgerAccountTypeConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetLedgerAccountTypeConfig(ctx, sqlcdb.GetLedgerAccountTypeConfigParams{
		MerchantID: merchantID,
		Type:       t,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *sqlxConfigRepo) UpsertLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType, accountID string, updatedBy *string) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	return r.q.UpsertLedgerAccountTypeConfig(ctx, sqlcdb.UpsertLedgerAccountTypeConfigParams{
		MerchantID: merchantID,
		Type:       t,
		AccountID:  accountID,
		UpdatedBy:  updatedBy,
	})
}

func (r *sqlxConfigRepo) GetAssetTypeAccountConfigs(ctx context.Context) ([]sqlcdb.AssetTypeAccountConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	return r.q.GetAssetTypeAccountConfigs(ctx, merchantID)
}

func (r *sqlxConfigRepo) GetAssetTypeAccountConfig(ctx context.Context, at enums.AssetType) (*sqlcdb.AssetTypeAccountConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetAssetTypeAccountConfig(ctx, sqlcdb.GetAssetTypeAccountConfigParams{
		MerchantID: merchantID,
		AssetType:  at,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *sqlxConfigRepo) UpsertAssetTypeAccountConfig(ctx context.Context, params sqlcdb.UpsertAssetTypeAccountConfigParams) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}
	params.MerchantID = merchantID
	return r.q.UpsertAssetTypeAccountConfig(ctx, params)
}

func (r *sqlxConfigRepo) GetAccountDescendants(ctx context.Context, ancestorID string) ([]projection.Account, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAccountDescendants(ctx, sqlcdb.GetAccountDescendantsParams{
		AncestorID: ancestorID,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.Account{
			MerchantID:       merchantID,
			AccountId:        row.AccountID,
			ParentId:         row.ParentID,
			Name:             row.Name,
			Type:             row.Type,
			NormalBalance:    row.NormalBalance,
			Currency:         row.Currency,
			IsSummary:        row.IsSummary,
			IsActive:         row.IsActive,
			Note:             row.Note,
			Version:          row.Version,
			HasChild:         row.HasChild,
			CashFlowCategory: dbmapconv.Ptr(row.CashFlowCategory),
			UpdatedBy:        row.UpdatedBy,
			UpdatedAt:        row.UpdatedAt,
		}
	}
	return result, nil
}
