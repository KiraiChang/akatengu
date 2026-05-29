package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type InvestmentRepo interface {
	GetByID(ctx context.Context, id int64) (*projection.Investment, error)
	GetByCreationEventUuid(ctx context.Context, uuid string) (*projection.Investment, error)
	GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error)
	GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error)
	GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error)
	GetNotCloseLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error)
	GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error)
	GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error)
	GetInvestmentPaged(ctx context.Context, req model.PaginationParams) ([]projection.Investment, int64, error)
	GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error)
	GetOpenLotsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLot, int64, error)
	GetMovementPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentMovement, int64, error)
}

type sqlcdbInvestmentRepository struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func (r *sqlcdbInvestmentRepository) GetMovementPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentMovement, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetInvestmentMovementsPaged(ctx, sqlcdb.GetInvestmentMovementsPagedParams{
		MerchantID:   merchantID,
		InvestmentID: id,
		Offset:       req.Offset,
		Limit:        req.Limit,
	})
	if err != nil {
		return nil, 0, err
	}
	total := int64(0)
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.InvestmentMovement, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentMovementFromGetInvestmentMovementsPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlcdbInvestmentRepository) GetLotDisposalsPaged(ctx context.Context, req model.PaginationParams, id int64) ([]projection.InvestmentLotDisposals, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetOpenLotDisposalsPaged(ctx, sqlcdb.GetOpenLotDisposalsPagedParams{
		MerchantID: merchantID,
		LotID:      &id,
		Offset:     req.Offset,
		Limit:      req.Limit,
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetOpenLotsPaged(ctx, sqlcdb.GetOpenLotsPagedParams{
		MerchantID:   merchantID,
		InvestmentID: id,
		Offset:       req.Offset,
		Limit:        req.Limit,
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q.GetInvestmentPaged(ctx, sqlcdb.GetInvestmentPagedParams{
		MerchantID: merchantID,
		Offset:     req.Offset,
		Limit:      req.Limit,
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := r.q.GetInvestment(ctx, sqlcdb.GetInvestmentParams{
		InvestmentID: id,
		MerchantID:   merchantID,
	})
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPtrFromGetInvestmentRow(result), nil
}

func (r *sqlcdbInvestmentRepository) GetByCreationEventUuid(ctx context.Context, uuid string) (*projection.Investment, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := r.q.GetInvestmentByUuid(ctx, sqlcdb.GetInvestmentByUuidParams{
		Uuid:       uuid,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPtrFromGetInvestmentByUuidRow(result), nil
}

func (r *sqlcdbInvestmentRepository) GetBySymbol(ctx context.Context, symbol, currency string) (*projection.Investment, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := r.q.GetInvestmentBySymbol(ctx, sqlcdb.GetInvestmentBySymbolParams{
		Symbol:     symbol,
		Currency:   currency,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPtrFromGetInvestmentBySymbolRow(result), nil
}

// querySummary: v_investment_summary 不支援 merchant_id 參數過濾，改以 sqlx 直接查詢底層表。
const querySummary = `
SELECT
    i.investment_id, i.symbol, i.name, i.asset_type, i.currency, i.cost_method,
    COALESCE(SUM(l.remaining_qty), 0) AS total_qty,
    CASE
        WHEN COALESCE(SUM(l.remaining_qty), 0) = 0 THEN 0
        ELSE COALESCE(SUM(l.remaining_qty * l.unit_cost), 0) / SUM(l.remaining_qty)
    END AS avg_cost_twd,
    COALESCE(SUM(l.remaining_qty * l.unit_cost), 0) AS total_cost_twd
FROM investments i
    LEFT JOIN investment_lots l ON i.investment_id = l.investment_id AND l.status != 'CLOSED'
WHERE i.is_active = 1 AND i.merchant_id = ? AND i.investment_id = ?
GROUP BY i.investment_id`

const queryAllSummaries = `
SELECT
    i.investment_id, i.symbol, i.name, i.asset_type, i.currency, i.cost_method,
    COALESCE(SUM(l.remaining_qty), 0) AS total_qty,
    CASE
        WHEN COALESCE(SUM(l.remaining_qty), 0) = 0 THEN 0
        ELSE COALESCE(SUM(l.remaining_qty * l.unit_cost), 0) / SUM(l.remaining_qty)
    END AS avg_cost_twd,
    COALESCE(SUM(l.remaining_qty * l.unit_cost), 0) AS total_cost_twd
FROM investments i
    LEFT JOIN investment_lots l ON i.investment_id = l.investment_id AND l.status != 'CLOSED'
WHERE i.is_active = 1 AND i.merchant_id = ?
GROUP BY i.investment_id`

func (r *sqlcdbInvestmentRepository) GetSummary(ctx context.Context, id int64) (*projection.InvestmentSummary, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	var s projection.InvestmentSummary
	err = r.db.QueryRowxContext(ctx, querySummary, merchantID, id).StructScan(&s)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sqlcdbInvestmentRepository) GetAllSummaries(ctx context.Context) ([]projection.InvestmentSummary, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	var summaries []projection.InvestmentSummary
	err = r.db.SelectContext(ctx, &summaries, queryAllSummaries, merchantID)
	return summaries, err
}

func (r *sqlcdbInvestmentRepository) GetNotCloseLots(ctx context.Context, investmentID int64) ([]projection.InvestmentLot, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetNotStatusLots(ctx, sqlcdb.GetNotStatusLotsParams{
		MerchantID:   merchantID,
		InvestmentID: investmentID,
		Status:       enums.LotStatusClose.Enum(),
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.InvestmentLot, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentLotFromGetNotStatusLotsRow(row)
	}
	return result, nil
}

func (r *sqlcdbInvestmentRepository) GetPosition(ctx context.Context, id int64) (*projection.InvestmentPosition, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetPosition(ctx, sqlcdb.GetPositionParams{
		InvestmentID: id,
		MerchantID:   merchantID,
	})
	if err != nil {
		return nil, err
	}
	return projection.InvestmentPositionPtrFromGetPositionRow(row), nil
}

func (r *sqlcdbInvestmentRepository) GetMovements(ctx context.Context, investmentID int64) ([]projection.InvestmentMovement, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetInvestmentMovements(ctx, sqlcdb.GetInvestmentMovementsParams{
		InvestmentID: investmentID,
		MerchantID:   merchantID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]projection.InvestmentMovement, len(rows))
	for i, row := range rows {
		result[i] = projection.InvestmentMovementFromGetInvestmentMovementsRow(row)
	}
	return result, nil
}
