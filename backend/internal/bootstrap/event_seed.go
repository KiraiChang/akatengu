package bootstrap

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	seedMerchantID  = int64(1)
	seedAggregateID = "1"
)

// BootstrapEventSeeds 確保 merchant 1 的 seed 資料有對應的事件記錄，
// 使全量 replay 能從事件重建所有 projection 表，不依賴 SQL seed 直接寫入。
// 各 aggregate 獨立檢查：已有事件時跳過（冪等），僅在事件不存在時才建立。
func BootstrapEventSeeds(ctx context.Context, db *sqlx.DB, es *services.EventStoreService, logger *zap.Logger) error {
	ctx = context.WithValue(ctx, ctxkey.MerchantID, seedMerchantID)
	eventRepo := query.NewEventRepo(db)

	if err := bootstrapAccounts(ctx, db, es, eventRepo, logger); err != nil {
		return fmt.Errorf("event seed accounts: %w", err)
	}
	if err := bootstrapSysAccounts(ctx, db, es, eventRepo, logger); err != nil {
		return fmt.Errorf("event seed sys_accounts: %w", err)
	}
	if err := bootstrapAccountConfigs(ctx, db, es, eventRepo, logger); err != nil {
		return fmt.Errorf("event seed account_configs: %w", err)
	}

	logger.Info("event seed bootstrap complete")
	return nil
}

// bootstrapAccounts 若 ACCOUNT aggregate 尚無事件，以當前 accounts 表建立一筆 AccountBulkImported 事件。
func bootstrapAccounts(ctx context.Context, db *sqlx.DB, es *services.EventStoreService, eventRepo query.EventRepo, logger *zap.Logger) error {
	aggType := enums.AggregateAccount.Enum()
	events, err := eventRepo.GetByAggregate(ctx, aggType, seedAggregateID)
	if err != nil {
		return err
	}
	if len(events) > 0 {
		return nil
	}

	// sqlx direct: bootstrap reads raw DB state to build initial event payload;
	// no sqlc query covers this cross-concern startup scan.
	type accountRow struct {
		AccountId        string  `db:"account_id"`
		ParentId         *string `db:"parent_id"`
		Name             string  `db:"name"`
		Type             string  `db:"type"`
		NormalBalance    string  `db:"normal_balance"`
		Currency         string  `db:"currency"`
		IsSummary        bool    `db:"is_summary"`
		IsActive         bool    `db:"is_active"`
		Note             *string `db:"note"`
		CashFlowCategory *string `db:"cash_flow_category"`
	}
	var rows []accountRow
	const q = `SELECT account_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, note, cash_flow_category FROM accounts WHERE merchant_id = ? ORDER BY account_id ASC`
	if err := db.SelectContext(ctx, &rows, q, seedMerchantID); err != nil {
		return fmt.Errorf("read accounts: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	accounts := make([]payload.AccountCreatePayload, len(rows))
	for i, r := range rows {
		accountType, err := enums.ParseAccountType(r.Type)
		if err != nil {
			return fmt.Errorf("parse account_type %q: %w", r.Type, err)
		}
		normalBalance, err := enums.ParseNormalBalance(r.NormalBalance)
		if err != nil {
			return fmt.Errorf("parse normal_balance %q: %w", r.NormalBalance, err)
		}
		var cashFlowCat *enums.CashFlowCategory
		if r.CashFlowCategory != nil {
			cat, err := enums.ParseCashFlowCategory(*r.CashFlowCategory)
			if err != nil {
				return fmt.Errorf("parse cash_flow_category %q: %w", *r.CashFlowCategory, err)
			}
			cashFlowCat = &cat
		}
		accounts[i] = payload.AccountCreatePayload{
			AccountId:        r.AccountId,
			ParentId:         r.ParentId,
			Name:             r.Name,
			Type:             accountType,
			NormalBalance:    normalBalance,
			Currency:         r.Currency,
			IsSummary:        r.IsSummary,
			IsActive:         r.IsActive,
			Note:             r.Note,
			CashFlowCategory: cashFlowCat,
		}
	}

	b, err := json.Marshal(payload.AccountBulkImportedPayload{Accounts: accounts})
	if err != nil {
		return err
	}
	if _, err := es.Append(ctx, cmd.AppendCmd{
		AggregateType:   aggType,
		AggregateID:     seedAggregateID,
		ExpectedVersion: 0,
		EventType:       event_types.EventAccountBulkImported.Enum(),
		Payload:         b,
	}); err != nil {
		return fmt.Errorf("append account.bulk_imported: %w", err)
	}
	logger.Info("event seed: AccountBulkImported created", zap.Int("count", len(accounts)))
	return nil
}

// bootstrapSysAccounts 若 SYS_CONFIG aggregate 尚無事件，為每筆 sys_accounts 各建立一筆 SysAccountUpdated 事件。
func bootstrapSysAccounts(ctx context.Context, db *sqlx.DB, es *services.EventStoreService, eventRepo query.EventRepo, logger *zap.Logger) error {
	aggType := enums.AggregateSysConfig.Enum()
	events, err := eventRepo.GetByAggregate(ctx, aggType, seedAggregateID)
	if err != nil {
		return err
	}
	if len(events) > 0 {
		return nil
	}

	// sqlx direct: bootstrap reads raw DB state to build initial event payloads.
	type sysRow struct {
		SysCode     string `db:"sys_code"`
		Description string `db:"description"`
		AccountID   string `db:"account_id"`
	}
	var rows []sysRow
	const q = `SELECT sys_code, description, account_id FROM sys_accounts WHERE merchant_id = ? ORDER BY sys_code ASC`
	if err := db.SelectContext(ctx, &rows, q, seedMerchantID); err != nil {
		return fmt.Errorf("read sys_accounts: %w", err)
	}
	if len(rows) == 0 {
		return nil
	}

	evtType := event_types.EventSysAccountUpdated.Enum()
	for i, r := range rows {
		b, err := json.Marshal(payload.SysAccountUpdatedPayload{
			SysCode:     r.SysCode,
			Description: r.Description,
			AccountID:   r.AccountID,
		})
		if err != nil {
			return err
		}
		if _, err := es.Append(ctx, cmd.AppendCmd{
			AggregateType:   aggType,
			AggregateID:     seedAggregateID,
			ExpectedVersion: int64(i),
			EventType:       evtType,
			Payload:         b,
		}); err != nil {
			return fmt.Errorf("append sys_config.sys_account_updated (%s): %w", r.SysCode, err)
		}
	}
	logger.Info("event seed: SysAccountUpdated created", zap.Int("count", len(rows)))
	return nil
}

// bootstrapAccountConfigs 若 ACCOUNT_CONFIG aggregate 尚無事件，為 ledger 類型與 asset 類型設定各建立對應事件。
func bootstrapAccountConfigs(ctx context.Context, db *sqlx.DB, es *services.EventStoreService, eventRepo query.EventRepo, logger *zap.Logger) error {
	aggType := enums.AggregateAccountConfig.Enum()
	events, err := eventRepo.GetByAggregate(ctx, aggType, seedAggregateID)
	if err != nil {
		return err
	}
	if len(events) > 0 {
		return nil
	}

	version := int64(0)

	// Ledger account type configs
	// sqlx direct: bootstrap reads raw DB state to build initial event payloads.
	type ledgerRow struct {
		Type      string `db:"type"`
		AccountID string `db:"account_id"`
	}
	var ledgerRows []ledgerRow
	const lq = `SELECT type, account_id FROM ledger_account_type_config WHERE merchant_id = ? ORDER BY type ASC`
	if err := db.SelectContext(ctx, &ledgerRows, lq, seedMerchantID); err != nil {
		return fmt.Errorf("read ledger_account_type_config: %w", err)
	}
	ledgerEvtType := event_types.EventLedgerAccountTypeConfigUpdated.Enum()
	for _, r := range ledgerRows {
		t, err := enums.ParseLedgerAccountType(r.Type)
		if err != nil {
			return fmt.Errorf("parse ledger_account_type %q: %w", r.Type, err)
		}
		b, err := json.Marshal(payload.LedgerAccountTypeConfigUpdatedPayload{
			Type:      t,
			AccountID: r.AccountID,
		})
		if err != nil {
			return err
		}
		if _, err := es.Append(ctx, cmd.AppendCmd{
			AggregateType:   aggType,
			AggregateID:     seedAggregateID,
			ExpectedVersion: version,
			EventType:       ledgerEvtType,
			Payload:         b,
		}); err != nil {
			return fmt.Errorf("append account_config.ledger_type_updated (%s): %w", r.Type, err)
		}
		version++
	}

	// Asset type account configs
	type assetRow struct {
		AssetType               string  `db:"asset_type"`
		RealizedGainAccountID   string  `db:"realized_gain_account_id"`
		RealizedLossAccountID   string  `db:"realized_loss_account_id"`
		UnrealizedGainAccountID string  `db:"unrealized_gain_account_id"`
		UnrealizedLossAccountID string  `db:"unrealized_loss_account_id"`
		OCIAccountID            *string `db:"oci_account_id"`
		FeeAccountID            string  `db:"fee_account_id"`
		TaxAccountID            string  `db:"tax_account_id"`
		AccountID               *string `db:"account_id"`
	}
	var assetRows []assetRow
	const aq = `SELECT asset_type, realized_gain_account_id, realized_loss_account_id, unrealized_gain_account_id, unrealized_loss_account_id, oci_account_id, fee_account_id, tax_account_id, account_id FROM asset_type_account_config WHERE merchant_id = ? ORDER BY asset_type ASC`
	if err := db.SelectContext(ctx, &assetRows, aq, seedMerchantID); err != nil {
		return fmt.Errorf("read asset_type_account_config: %w", err)
	}
	assetEvtType := event_types.EventAssetTypeAccountConfigUpdated.Enum()
	for _, r := range assetRows {
		at, err := enums.ParseAssetType(r.AssetType)
		if err != nil {
			return fmt.Errorf("parse asset_type %q: %w", r.AssetType, err)
		}
		b, err := json.Marshal(payload.AssetTypeAccountConfigUpdatedPayload{
			AssetType:               at,
			RealizedGainAccountID:   r.RealizedGainAccountID,
			RealizedLossAccountID:   r.RealizedLossAccountID,
			UnrealizedGainAccountID: r.UnrealizedGainAccountID,
			UnrealizedLossAccountID: r.UnrealizedLossAccountID,
			OCIAccountID:            r.OCIAccountID,
			FeeAccountID:            r.FeeAccountID,
			TaxAccountID:            r.TaxAccountID,
			AccountID:               r.AccountID,
		})
		if err != nil {
			return err
		}
		if _, err := es.Append(ctx, cmd.AppendCmd{
			AggregateType:   aggType,
			AggregateID:     seedAggregateID,
			ExpectedVersion: version,
			EventType:       assetEvtType,
			Payload:         b,
		}); err != nil {
			return fmt.Errorf("append account_config.asset_type_updated (%s): %w", r.AssetType, err)
		}
		version++
	}

	logger.Info("event seed: AccountConfig created",
		zap.Int("ledger_types", len(ledgerRows)),
		zap.Int("asset_types", len(assetRows)),
	)
	return nil
}
