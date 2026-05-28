package services

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type SettingService interface {
	GetLedgerAccountTypeConfigs(ctx context.Context) ([]projection.LedgerAccountTypeConfigResult, error)
	GetAssetTypeAccountConfigs(ctx context.Context) ([]projection.AssetTypeAccountConfigResult, error)
}

func NewSettingService(db *sqlx.DB) SettingService {
	return &settingService{
		r: query.NewConfigRepo(db),
	}
}

type settingService struct {
	r query.ConfigRepo
}

func (s *settingService) GetLedgerAccountTypeConfigs(ctx context.Context) ([]projection.LedgerAccountTypeConfigResult, error) {
	configs, err := s.r.GetLedgerAccountTypeConfigs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccountTypeConfigResult, len(configs))
	for i, cfg := range configs {
		descendants, err := s.r.GetAccountDescendants(ctx, cfg.AccountID)
		if err != nil {
			return nil, err
		}
		result[i] = projection.LedgerAccountTypeConfigResult{
			LedgerAccountTypeConfig: cfg,
			Descendants:             descendants,
		}
	}
	return result, nil
}

func (s *settingService) GetAssetTypeAccountConfigs(ctx context.Context) ([]projection.AssetTypeAccountConfigResult, error) {
	configs, err := s.r.GetAssetTypeAccountConfigs(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.AssetTypeAccountConfigResult, len(configs))
	for i, cfg := range configs {
		var descendants []projection.Account
		if cfg.AccountID != nil {
			descendants, err = s.r.GetAccountDescendants(ctx, *cfg.AccountID)
			if err != nil {
				return nil, err
			}
		} else {
			descendants = []projection.Account{}
		}
		result[i] = projection.AssetTypeAccountConfigResult{
			AssetTypeAccountConfig: cfg,
			Descendants:            descendants,
		}
	}
	return result, nil
}

