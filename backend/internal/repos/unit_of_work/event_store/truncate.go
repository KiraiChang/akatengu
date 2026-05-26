package event_store

import (
	"akatengu/internal/pkg/ctxkey"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type sqlxTruncateRepository struct {
	tx *sqlx.Tx
}

// TruncateProjections 清除當前商戶的 projection 資料，供全量重播前使用。
// 清除順序依 FK 依賴由子到父；accounts 使用遞迴 CTE 保留 sys_accounts 祖先鏈。
// 直接使用 sqlx 原因：批次 15 張表的 DELETE，若各自定義 sqlc query 無實質效益。
func (r *sqlxTruncateRepository) TruncateProjections(ctx context.Context) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return err
	}

	// 依 FK 依賴由子到父依序刪除，各表以 merchant_id 限定範圍。
	// exchange_rates 為全域資料（無 merchant_id 欄位），不在清除範圍；replay 時以 UPSERT 補回。
	stmts := []string{
		`DELETE FROM prepaid_amortizations WHERE merchant_id = ?`,
		`DELETE FROM prepaids WHERE merchant_id = ?`,
		`DELETE FROM fixed_asset_depreciations WHERE merchant_id = ?`,
		`DELETE FROM fixed_assets WHERE merchant_id = ?`,
		`DELETE FROM investment_lot_disposals WHERE merchant_id = ?`,
		`DELETE FROM journal_entries WHERE merchant_id = ?`,
		`DELETE FROM installment_payments WHERE merchant_id = ?`,
		`DELETE FROM installments WHERE merchant_id = ?`,
		`DELETE FROM investment_positions WHERE merchant_id = ?`,
		`DELETE FROM investment_lots WHERE merchant_id = ?`,
		`DELETE FROM investment_movements WHERE merchant_id = ?`,
		`DELETE FROM transactions WHERE merchant_id = ?`,
		`DELETE FROM investments WHERE merchant_id = ?`,
		`DELETE FROM period_closings WHERE merchant_id = ?`,
		`DELETE FROM ledger_accounts WHERE merchant_id = ?`,
		`DELETE FROM aggregate_versions WHERE merchant_id = ?`,
		`DELETE FROM snapshots WHERE merchant_id = ?`,
	}
	for _, stmt := range stmts {
		if _, err := r.tx.ExecContext(ctx, stmt, merchantID); err != nil {
			return fmt.Errorf("truncate projections: %w", err)
		}
	}

	// 啟用 defer_foreign_keys，讓 accounts 自我參照（parent_id）的刪除可以在
	// commit 時才統一做 FK 驗證，避免子→父刪除順序問題。
	if _, err := r.tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		return fmt.Errorf("defer_foreign_keys: %w", err)
	}

	// account_closure 參照 accounts，先於 accounts 刪除。
	if _, err := r.tx.ExecContext(ctx, `DELETE FROM account_closure WHERE merchant_id = ?`, merchantID); err != nil {
		return fmt.Errorf("delete account_closure: %w", err)
	}

	// 只保留 sys_accounts 所依賴的祖先鏈，刪除當前商戶使用者建立的 accounts。
	// 遞迴 CTE 起點、JOIN、外層 DELETE 均以 merchant_id 限定，避免跨商戶誤刪。
	const deleteUserAccounts = `
		DELETE FROM accounts
		WHERE merchant_id = ?
		AND account_id NOT IN (
			WITH RECURSIVE protected(id) AS (
				SELECT account_id FROM sys_accounts WHERE merchant_id = ?
				UNION
				SELECT a.parent_id
				FROM   accounts a
				INNER  JOIN protected p ON a.account_id = p.id
				WHERE  a.parent_id IS NOT NULL AND a.merchant_id = ?
			)
			SELECT id FROM protected
		)`
	if _, err := r.tx.ExecContext(ctx, deleteUserAccounts, merchantID, merchantID, merchantID); err != nil {
		return fmt.Errorf("delete user accounts: %w", err)
	}

	// sqlite_sequence reset 在多租戶下不適用：其他商戶資料仍存在，
	// 重置後 AUTOINCREMENT 計數器歸零會與既有 row ID 衝突。
	// replay 以事件 payload 中的明確 ID 寫入，不依賴 sequence 從 1 開始。

	return nil
}