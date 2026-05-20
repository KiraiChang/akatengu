package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/pkg/dbmapconv"
	"context"

	"github.com/jmoiron/sqlx"
)

type ConfigRepo interface {
	GetLedgerAccountTypeConfigs(ctx context.Context) ([]projection.LedgerAccountTypeConfig, error)
	GetLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType) (*projection.LedgerAccountTypeConfig, error)
	UpsertLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType, accountID string, updatedBy *string) error

	GetAssetTypeAccountConfigs(ctx context.Context) ([]projection.AssetTypeAccountConfig, error)
	GetAssetTypeAccountConfig(ctx context.Context, at enums.AssetType) (*projection.AssetTypeAccountConfig, error)
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

func (r *sqlxConfigRepo) GetLedgerAccountTypeConfigs(ctx context.Context) ([]projection.LedgerAccountTypeConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetLedgerAccountTypeConfigs(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccountTypeConfig, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountTypeConfigFromLedgerAccountTypeConfig(row)
	}
	return result, nil
}

func (r *sqlxConfigRepo) GetLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType) (*projection.LedgerAccountTypeConfig, error) {
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
	return projection.LedgerAccountTypeConfigPtrFromLedgerAccountTypeConfig(row), nil
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

func (r *sqlxConfigRepo) GetAssetTypeAccountConfigs(ctx context.Context) ([]projection.AssetTypeAccountConfig, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAssetTypeAccountConfigs(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.AssetTypeAccountConfig, len(rows))
	for i, row := range rows {
		result[i] = projection.AssetTypeAccountConfigFromAssetTypeAccountConfig(row)
	}
	return result, nil
}

func (r *sqlxConfigRepo) GetAssetTypeAccountConfig(ctx context.Context, at enums.AssetType) (*projection.AssetTypeAccountConfig, error) {
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
	return projection.AssetTypeAccountConfigPtrFromAssetTypeAccountConfig(row), nil
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
