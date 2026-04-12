package projection_repo

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

type sqlxInvestmentRepo struct {
	tx *sqlx.Tx
}

func (r *sqlxInvestmentRepo) PositionSplit(ctx context.Context, id int64, ratio decimal.Decimal) error {
	query := `
		UPDATE investment_positions
		SET total_quantity = total_quantity * :ratio
		WHERE investment_id = :id
	`

	args := map[string]interface{}{
		"id":    id,
		"ratio": ratio,
	}

	_, err := r.tx.NamedExecContext(ctx, query, args)
	return err
}

func (r *sqlxInvestmentRepo) LotSplit(ctx context.Context, id int64, ratio decimal.Decimal) error {
	query := `
		UPDATE investment_lots
		SET quantity = quantity * :ratio,
		    remaining_qty = remaining_qty * :ratio,
			unit_cost = unit_cost / :ratio
		WHERE investment_id = :id
			AND status  <> 'CLOSED'
	`

	args := map[string]interface{}{
		"id":    id,
		"ratio": ratio,
	}

	_, err := r.tx.NamedExecContext(ctx, query, args)
	return err
}

func (r *sqlxInvestmentRepo) UpsertPosition(ctx context.Context, position projection.InvestmentPosition) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO investment_positions 
                (investment_id, total_quantity, total_cost)  
            VALUES (:investment_id, :total_quantity, :total_cost)
            ON CONFLICT (investment_id) 
			DO UPDATE
				SET total_quantity = total_quantity + excluded.total_cost,
					total_cost = total_cost + excluded.total_cost`,
		position,
	)
	return err
}

func (r *sqlxInvestmentRepo) InsertLot(ctx context.Context, lot projection.InvestmentLot) (int64, error) {
	result, err := r.tx.NamedExecContext(ctx, `
		INSERT INTO investment_lots
			(investment_id, acquired_date, movement_id,
			 quantity, unit_cost, total_cost, remaining_qty)
		VALUES (:investment_id, :acquired_date, :movement_id,
			 :quantity, :unit_cost, :total_cost, :remaining_qty)`,
		lot,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqlxInvestmentRepo) UpdateLot(ctx context.Context, id int64, txn_id int64) error {
	_, err := r.tx.ExecContext(ctx, `
		UPDATE investment_lots
		SET txn_id = ?
		WHERE lot_id = ?`,
		txn_id, id,
	)
	return err
}

func (r *sqlxInvestmentRepo) InsertMovement(ctx context.Context, m projection.InvestmentMovement) (int64, error) {
	result, err := r.tx.NamedExecContext(ctx, `
		INSERT INTO investment_movements
			(investment_id, movement_type, movement_date, event_id,
			 quantity, unit_price, unit_price_twd, exchange_rate,
			 fee, tax, realized_gain, cost_basis, split_ratio, gross_amount, net_amount, withholding_tax)
		VALUES (:investment_id, :movement_type, :movement_date, :event_id,
			 :quantity, :unit_price, :unit_price_twd, :exchange_rate,
			 :fee, :tax, :realized_gain, :cost_basis, :split_ratio, :gross_amount, :net_amount, :withholding_tax)`,
		m,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqlxInvestmentRepo) UpdateMovement(ctx context.Context, id int64, txn_id int64) error {
	_, err := r.tx.ExecContext(ctx, `
		UPDATE investment_movements
			SET txn_id = ?
		WHERE movement_id = ?`,
		txn_id, id,
	)
	return err
}

func (r *sqlxInvestmentRepo) InsertLotDisposals(ctx context.Context, disposals projection.InvestmentLotDisposals) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO investment_lot_disposals 
                (lot_id, movement_id, quantity, cost_basis, sale_proceeds, capital_gain, holding_period_days, disposal_date)
            VALUES (:lot_id, :movement_id, :quantity, :cost_basis, :sale_proceeds, :capital_gain, :holding_period_days, :disposal_date)`,
		disposals,
	)
	return err
}

func (r *sqlxInvestmentRepo) UpsertExchangeRate(ctx context.Context, rate projection.ExchangeRate) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO exchange_rates
                (currency, rate_date, rate_twd, source)
            VALUES (:currency, :rate_date, :rate_twd, :source)
            ON CONFLICT (currency, rate_date) 
            DO UPDATE SET
            rate_twd = :rate_twd,
            source = :source`,
		rate,
	)
	return err
}

func (r *sqlxInvestmentRepo) CreateInvestment(ctx context.Context, p projection.Investment) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO investments
                (account_id, asset_type, currency, symbol, name, cost_method, is_active, version)
            VALUES (:account_id, :asset_type, :currency, :symbol, :name, :cost_method, :is_active, 1)`,
		p,
	)
	return err
}

func (r *sqlxInvestmentRepo) UpdateInvestment(ctx context.Context, p projection.Investment) error {
	_, err := r.tx.NamedExecContext(ctx, `
            UPDATE investments
                SET account_id = :account_id, 
                    asset_type = :asset_type, 
                    currency = :currency,
                    symbol = :symbol,
                    name = :name,
                    cost_method = :cost_method,
                    is_active = :is_active,
                    version = version + 1
            WHERE investment_id = :investment_id
            	AND version = :version`,
		p,
	)
	return err
}

func NewInvestmentRepo(tx *sqlx.Tx) InvestmentRepo {
	return &sqlxInvestmentRepo{tx: tx}
}
