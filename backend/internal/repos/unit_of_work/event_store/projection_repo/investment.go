package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/shopspring/decimal"
)

type sqlxInvestmentRepo struct {
	q *sqlcdb.Queries
}

func (r *sqlxInvestmentRepo) UpdateInvestmentPositionSold(ctx context.Context, p projection.InvestmentPosition) error {
	return r.q.UpdateInvestmentPositionSold(ctx, p.ToUpdateInvestmentPositionSoldParams())
}

func (r *sqlxInvestmentRepo) PositionSplit(ctx context.Context, id int64, ratio decimal.Decimal) error {
	return r.q.UpdateInvestmentPositionSplit(ctx, sqlcdb.UpdateInvestmentPositionSplitParams{
		InvestmentID:  id,
		TotalQuantity: ratio,
	})
}

func (r *sqlxInvestmentRepo) LotSplit(ctx context.Context, id int64, ratio decimal.Decimal) error {
	return r.q.UpdateInvestmentLotSplit(ctx, sqlcdb.UpdateInvestmentLotSplitParams{
		InvestmentID: id,
		Quantity:     ratio,
		RemainingQty: ratio,
		UnitCost:     ratio,
	})
}

func (r *sqlxInvestmentRepo) UpsertPosition(ctx context.Context, p projection.InvestmentPosition) error {
	return r.q.UpsertInvestmentPosition(ctx, p.ToUpsertInvestmentPositionParams())
}

func (r *sqlxInvestmentRepo) InsertLot(ctx context.Context, l projection.InvestmentLot) (int64, error) {
	return r.q.InsertInvestmentLot(ctx, l.ToInsertInvestmentLotParams())
}

func (r *sqlxInvestmentRepo) UpdateLot(ctx context.Context, id int64, txn_id int64) error {
	return r.q.UpdateInvestmentLotTxn(ctx, sqlcdb.UpdateInvestmentLotTxnParams{
		TxnID: &txn_id,
		LotID: id,
	})
}

func (r *sqlxInvestmentRepo) InsertMovement(ctx context.Context, m projection.InvestmentMovement) (int64, error) {
	return r.q.InsertInvestmentMovement(ctx, m.ToInsertInvestmentMovementParams())
}

func (r *sqlxInvestmentRepo) UpdateMovement(ctx context.Context, id int64, txn_id int64) error {
	return r.q.UpdateInvestmentMovementTxn(ctx, sqlcdb.UpdateInvestmentMovementTxnParams{
		TxnID:      &txn_id,
		MovementID: id,
	})
}

func (r *sqlxInvestmentRepo) InsertLotDisposals(ctx context.Context, d projection.InvestmentLotDisposals) error {
	return r.q.InsertInvestmentLotDisposal(ctx, d.ToInsertInvestmentLotDisposalParams())
}

func (r *sqlxInvestmentRepo) UpsertExchangeRate(ctx context.Context, rate projection.ExchangeRate) error {

	return r.q.UpsertExchangeRate(ctx, rate.ToUpsertExchangeRateParams())
}

func (r *sqlxInvestmentRepo) CreateInvestment(ctx context.Context, p projection.Investment) error {
	return r.q.CreateInvestment(ctx, p.ToCreateInvestmentParams())
}

func (r *sqlxInvestmentRepo) UpdateInvestment(ctx context.Context, p projection.Investment) error {
	return r.q.UpdateInvestment(ctx, p.ToUpdateInvestmentParams())
}

func NewInvestmentRepo(q *sqlcdb.Queries) InvestmentRepo {
	return &sqlxInvestmentRepo{q}
}
