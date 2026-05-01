package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InvestmentRepo interface {
	GetByID(ctx context.Context, id int64) (*projection.Investment, error)
	GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error)
	GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error)
	GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error)
	GetOpenLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error)
	GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error)
	GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error)
	GetInvestmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Investment, int64, error)
	GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error)
	GetOpenLotsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLot, int64, error)
}

type sqlcdbInvestmentRepository struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func (r *sqlcdbInvestmentRepository) GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error) {
	rows, err := r.q.GetOpenLotDisposalsPaged(ctx, sqlcdb.GetOpenLotDisposalsPagedParams{
		Offset: req.Offset,
		Limit:  req.Limit,
		LotID:  &id,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.InvestmentLotDisposals, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentLotDisposalsFromGetOpenLotDisposalsPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlcdbInvestmentRepository) GetOpenLotsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLot, int64, error) {
	rows, err := r.q.GetOpenLotsPaged(ctx, sqlcdb.GetOpenLotsPagedParams{
		Offset:       req.Offset,
		Limit:        req.Limit,
		InvestmentID: id,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.InvestmentLot, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentLotFromGetOpenLotsPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlcdbInvestmentRepository) GetInvestmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Investment, int64, error) {
	rows, err := r.q.GetInvestmentPaged(ctx, sqlcdb.GetInvestmentPagedParams{
		Offset: req.Offset,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.Investment, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentFromGetInvestmentPagedRow(row)
	}
	return result, total, nil
}

func newInvestmentRepo(q *sqlcdb.Queries, db *sqlx.DB) InvestmentRepo {
	return &sqlcdbInvestmentRepository{q: q, db: db}
}

func NewInvestmentRepo(db *sqlx.DB) InvestmentRepo {
	return &sqlcdbInvestmentRepository{q: sqlcdb.New(db), db: db}
}

func (r *sqlcdbInvestmentRepository) GetByID(ctx context.Context, id int64) (*projection.Investment, error) {
	row, err := r.q.GetInvestment(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPtrFromInvestment(row), nil
}

func (r *sqlcdbInvestmentRepository) GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error) {
	row, err := r.q.GetInvestmentBySymbol(ctx, sqlcdb.GetInvestmentBySymbolParams{
		Symbol:   symbol,
		Currency: currency,
	})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPtrFromInvestment(row), nil
}

func (r *sqlcdbInvestmentRepository) GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error) {
	var s projection.InvestmentSummary
	err := r.db.QueryRowxContext(ctx,
		`SELECT investment_id, symbol, name, asset_type, currency,
		        cost_method, total_qty, avg_cost_twd, total_cost_twd
		 FROM v_investment_summary WHERE investment_id = ?`, id,
	).StructScan(&s)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sqlcdbInvestmentRepository) GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error) {
	var summaries []projection.InvestmentSummary
	err := r.db.SelectContext(ctx, &summaries,
		`SELECT investment_id, symbol, name, asset_type, currency,
		        cost_method, total_qty, avg_cost_twd, total_cost_twd
		 FROM v_investment_summary`)
	return summaries, err
}

func (r *sqlcdbInvestmentRepository) GetOpenLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error) {
	rows, err := r.q.GetOpenLots(ctx, sqlcdb.GetOpenLotsParams{
		InvestmentID: investmentID,
		Status:       enums.LotStatusClose.Enum(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.InvestmentLot, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentLotFromInvestmentLot(row)
	}
	return result, nil
}

func (r *sqlcdbInvestmentRepository) GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error) {
	row, err := r.q.GetInvestmentPosition(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPositionPtrFromInvestmentPosition(row), nil
}

func (r *sqlcdbInvestmentRepository) GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error) {
	rows, err := r.q.GetInvestmentMovements(ctx, investmentID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.InvestmentMovement, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentMovementFromGetInvestmentMovementsRow(row)
	}
	return result, nil
}
