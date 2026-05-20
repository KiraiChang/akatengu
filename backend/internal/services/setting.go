package services

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type LedgerAccountTypeConfigResult struct {
	sqlcdb.LedgerAccountTypeConfig
	Descendants []projection.Account `json:"descendants"`
}

type SettingService interface {
	GetLedgerAccountTypeConfigs(ctx context.Context) ([]LedgerAccountTypeConfigResult, error)
	UpdateLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType, accountID string, updatedBy *string) error

	GetAssetTypeAccountConfigs(ctx context.Context) ([]sqlcdb.AssetTypeAccountConfig, error)
	UpdateAssetTypeAccountConfig(ctx context.Context, params sqlcdb.UpsertAssetTypeAccountConfigParams) error
}

func NewSettingService(db *sqlx.DB) SettingService {
	return &settingService{
		r: query.NewConfigRepo(db),
	}
}

type settingService struct {
	r query.ConfigRepo
}

func (s *settingService) GetLedgerAccountTypeConfigs(ctx context.Context) ([]LedgerAccountTypeConfigResult, error) {
	configs, err := s.r.GetLedgerAccountTypeConfigs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]LedgerAccountTypeConfigResult, len(configs))
	for i, cfg := range configs {
		descendants, err := s.r.GetAccountDescendants(ctx, cfg.AccountID)
		if err != nil {
			return nil, err
		}
		result[i] = LedgerAccountTypeConfigResult{
			LedgerAccountTypeConfig: cfg,
			Descendants:             descendants,
		}
	}
	return result, nil
}

func (s *settingService) UpdateLedgerAccountTypeConfig(ctx context.Context, t enums.LedgerAccountType, accountID string, updatedBy *string) error {
	return s.r.UpsertLedgerAccountTypeConfig(ctx, t, accountID, updatedBy)
}

func (s *settingService) GetAssetTypeAccountConfigs(ctx context.Context) ([]sqlcdb.AssetTypeAccountConfig, error) {
	return s.r.GetAssetTypeAccountConfigs(ctx)
}

func (s *settingService) UpdateAssetTypeAccountConfig(ctx context.Context, params sqlcdb.UpsertAssetTypeAccountConfigParams) error {
	return s.r.UpsertAssetTypeAccountConfig(ctx, params)
}
