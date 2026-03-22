package query

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InvestmentRepo interface {
	// 查詢
	GetByID(ctx context.Context, id int64) (*projection.Investment, error)
	GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error)
	GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error)
	GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error)

	// FIFO：取得開放批次（依買入日期排序）
	GetOpenLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error)

	// 異動歷史
	GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error)
}

// ─────────────────────────────────────────
// sqlx 實作
// ─────────────────────────────────────────

type sqlxInvestmentRepository struct{ db *sqlx.DB }

func NewInvestmentRepo(db *sqlx.DB) InvestmentRepo {
	return &sqlxInvestmentRepository{db: db}
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
