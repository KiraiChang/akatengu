package projection_repo

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxInvestmentRepo struct {
	tx *sqlx.Tx
}

func NewInvestmentRepo(tx *sqlx.Tx) InvestmentRepo {
	return &sqlxInvestmentRepo{tx: tx}
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
