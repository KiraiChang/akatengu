package repos

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

// ─────────────────────────────────────────
// Interface
// ─────────────────────────────────────────

type InvestmentRepository interface {
	// 查詢
	GetByID(ctx context.Context, id int64) (*projection.Investment, error)
	GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error)
	GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error)
	GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error)

	// FIFO：取得開放批次（依買入日期排序）
	GetOpenLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error)

	// 異動歷史
	GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error)

	// 寫入（走 tx）
	WithTx(ctx context.Context, fn func(InvestmentTxRepository) error) error
}

type InvestmentTxRepository interface {
	InsertInvestment(ctx context.Context, inv projection.Investment) (int64, error)
	InsertLot(ctx context.Context, lot projection.InvestmentLot) (int64, error)
	UpdateLot(ctx context.Context, lotID int64, remainingQty decimal.Decimal, status enums.LotStatus) error
	InsertMovement(ctx context.Context, m projection.InvestmentMovement) (int64, error)

	// 平均成本法：更新既有批次的加權平均成本
	UpsertAvgLot(ctx context.Context, investmentID int64, lot projection.InvestmentLot) error
}

// ─────────────────────────────────────────
// sqlx 實作
// ─────────────────────────────────────────

type sqlxInvestmentRepository struct{ db *sqlx.DB }

func NewInvestmentRepository(db *sqlx.DB) InvestmentRepository {
	return &sqlxInvestmentRepository{db: db}
}

func (r *sqlxInvestmentRepository) WithTx(ctx context.Context, fn func(InvestmentTxRepository) error) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	if err := fn(&sqlxInvestmentTxRepository{tx: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *sqlxInvestmentRepository) GetByID(ctx context.Context, id int64) (*projection.Investment, error) {
	var inv projection.Investment
	err := r.db.QueryRowxContext(ctx, `SELECT * FROM investments WHERE investment_id = ?`, id).
		StructScan(&inv)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}

func (r *sqlxInvestmentRepository) GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error) {
	var inv projection.Investment
	err := r.db.QueryRowxContext(ctx,
		`SELECT * FROM investments WHERE symbol = ? AND currency = ?`,
		symbol, currency,
	).StructScan(&inv)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &inv, err
}

func (r *sqlxInvestmentRepository) GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error) {
	var s projection.InvestmentSummary
	err := r.db.QueryRowxContext(ctx,
		`SELECT * FROM v_investment_summary WHERE investment_id = ?`, id,
	).StructScan(&s)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *sqlxInvestmentRepository) GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error) {
	var summaries []projection.InvestmentSummary
	err := r.db.SelectContext(ctx, &summaries, `SELECT * FROM v_investment_summary`)
	return summaries, err
}

func (r *sqlxInvestmentRepository) GetOpenLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error) {
	var lots []projection.InvestmentLot
	err := r.db.SelectContext(ctx, &lots, `
		SELECT * FROM investment_lots
		WHERE investment_id = ?
		  AND status != ?
		ORDER BY acquired_date, lot_id`,
		investmentID, enums.LotStatusClose,
	)
	return lots, err
}

func (r *sqlxInvestmentRepository) GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error) {
	var movements []projection.InvestmentMovement
	err := r.db.SelectContext(ctx, &movements, `
		SELECT * FROM investment_movements
		WHERE investment_id = ?
		ORDER BY movement_date, movement_id`,
		investmentID,
	)
	return movements, err
}

// ─────────────────────────────────────────
// Tx 實作
// ─────────────────────────────────────────

type sqlxInvestmentTxRepository struct{ tx *sqlx.Tx }

func (r *sqlxInvestmentTxRepository) InsertInvestment(ctx context.Context, inv projection.Investment) (int64, error) {
	result, err := r.tx.ExecContext(ctx, `
		INSERT INTO investments
			(account_id, asset_type, currency, symbol, name, cost_method)
		VALUES (?, ?, ?, ?, ?, ?)`,
		inv.AccountId, inv.AssetType, inv.Currency, inv.Symbol, inv.Name, inv.CostMethod,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqlxInvestmentTxRepository) InsertLot(ctx context.Context, lot projection.InvestmentLot) (int64, error) {
	result, err := r.tx.ExecContext(ctx, `
		INSERT INTO investment_lots
			(investment_id, acquired_date, txn_id,
			 quantity, unit_cost, unit_cost_twd, remaining_qty, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		lot.InvestmentId, lot.AcquiredDate, lot.TransactionId,
		lot.Quantity, lot.UnitCost, lot.UnitCostTWD, lot.RemainingQty, lot.Status,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqlxInvestmentTxRepository) UpdateLot(ctx context.Context, lotID int64, remainingQty decimal.Decimal, status enums.LotStatus) error {
	_, err := r.tx.ExecContext(ctx, `
		UPDATE investment_lots
		SET remaining_qty = ?, status = ?
		WHERE lot_id = ?`,
		remainingQty, status, lotID,
	)
	return err
}

func (r *sqlxInvestmentTxRepository) InsertMovement(ctx context.Context, m projection.InvestmentMovement) (int64, error) {
	result, err := r.tx.ExecContext(ctx, `
		INSERT INTO investment_movements
			(investment_id, txn_id, movement_type, movement_date,
			 quantity, unit_price, unit_price_twd, exchange_rate,
			 fee_twd, tax_twd, realized_gain_twd, cost_basis_twd)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.InvestmentId, m.TransactionId, m.MovementType, m.MovementDate,
		m.Quantity, m.UnitPrice, m.UnitPriceTWD, m.ExchangeRate,
		m.FeeTWD, m.TaxTWD, m.RealizedGainTWD, m.CostBasisTWD,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpsertAvgLot 平均成本法：只保留一筆批次，每次買入重新計算加權平均
func (r *sqlxInvestmentTxRepository) UpsertAvgLot(ctx context.Context, investmentID int64, lot projection.InvestmentLot) error {
	_, err := r.tx.ExecContext(ctx, `
		INSERT INTO investment_lots
			(investment_id, acquired_date, txn_id,
			 quantity, unit_cost, unit_cost_twd, remaining_qty, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO NOTHING`,
		investmentID, lot.AcquiredDate, lot.TransactionId,
		lot.Quantity, lot.UnitCost, lot.UnitCostTWD, lot.RemainingQty, enums.LotStatusOpen,
	)
	return err
}
