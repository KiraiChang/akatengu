package event_store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type sqlxTruncateRepository struct {
	tx *sqlx.Tx
}

// TruncateProjections 清除所有 projection 資料，供全量重播前使用。
// 清除順序依 FK 依賴由子到父；accounts 使用遞迴 CTE 保留 sys_accounts 祖先鏈。
func (r *sqlxTruncateRepository) TruncateProjections(ctx context.Context) error {
	// 依 FK 依賴由子到父依序刪除
	stmts := []string{
		`DELETE FROM investment_lot_disposals`,
		`DELETE FROM journal_entries`,
		`DELETE FROM installment_payments`,
		`DELETE FROM installments`,
		`DELETE FROM investment_positions`,
		`DELETE FROM investment_lots`,
		`DELETE FROM investment_movements`,
		`DELETE FROM transactions`,
		`DELETE FROM investments`,
		`DELETE FROM period_closings`,
		`DELETE FROM exchange_rates`,
		`DELETE FROM ledger_accounts`,
	}
	for _, stmt := range stmts {
		if _, err := r.tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("truncate projections: %w", err)
		}
	}

	// 啟用 defer_foreign_keys，讓 accounts 自我參照（parent_id）的刪除可以在
	// commit 時才統一做 FK 驗證，避免子→父刪除順序問題。
	if _, err := r.tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		return fmt.Errorf("defer_foreign_keys: %w", err)
	}

	// 只保留 sys_accounts 所依賴的祖先鏈，刪除所有使用者建立的 accounts。
	// 遞迴 CTE 從 sys_accounts 直接引用的 account 向上追溯所有父科目，
	// 這些科目是 seed 資料的一部分，不應被重播覆蓋。
	const deleteUserAccounts = `
		DELETE FROM accounts
		WHERE account_id NOT IN (
			WITH RECURSIVE protected(id) AS (
				SELECT account_id FROM sys_accounts
				UNION
				SELECT a.parent_id
				FROM   accounts a
				INNER  JOIN protected p ON a.account_id = p.id
				WHERE  a.parent_id IS NOT NULL
			)
			SELECT id FROM protected
		)`
	if _, err := r.tx.ExecContext(ctx, deleteUserAccounts); err != nil {
		return fmt.Errorf("delete user accounts: %w", err)
	}

	// 重置 AUTOINCREMENT 序列，確保重播後的 ID 從 1 重新開始，
	// 以吻合事件 payload 中記錄的 ID（如 investment_id, installment_id）。
	const resetSeq = `
		DELETE FROM sqlite_sequence
		WHERE name IN (
			'installments', 'installment_payments',
			'investments',
			'investment_lots', 'investment_movements',
			'investment_positions', 'investment_lot_disposals',
			'period_closings', 'exchange_rates'
		)`
	if _, err := r.tx.ExecContext(ctx, resetSeq); err != nil {
		return fmt.Errorf("reset sqlite_sequence: %w", err)
	}

	return nil
}