package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"

	"github.com/shopspring/decimal"
)

type AccountBalanceRealtimeProjection struct{}

func (s *AccountBalanceRealtimeProjection) Name() string {
	return enums.ProjectionTypeAccountBalanceRealtime.String()
}

func (s *AccountBalanceRealtimeProjection) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.Val() {
	case event_types.EventTransactionCreated:
		return s.applyCreated(ctx, tx, ct)
	case event_types.EventTransactionVoided:
		return s.applyVoided(ctx, tx, ct)
	case event_types.EventTransactionCorrected:
		return s.applyCorrected(ctx, tx, ct)

	case event_types.EventInvestmentBought:
		return s.applyInvestmentBought(ctx, tx, ct)
	case event_types.EventInvestmentSold:
		return s.applyInvestmentSold(ctx, tx, ct)
	case event_types.EventDividendReceived:
		return s.applyDividendReceived(ctx, tx, ct)
	}
	return nil
}

func (s *AccountBalanceRealtimeProjection) applyCreated(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCreatedPayload](ct)
	if err != nil {
		return err
	}
	// void-reference transactions have reversed entries already captured by EventTransactionVoided
	if p.RefTxnId != nil {
		return nil
	}
	deltas := make([]balanceDelta, len(p.Entries))
	for i, e := range p.Entries {
		deltas[i] = balanceDelta{
			accountId: e.AccountId,
			ledgerId:  e.LedgerId,
			debit:     e.Debit,
			credit:    e.Credit,
		}
	}
	return s.applyDeltas(ctx, tx, ct.MerchantID, deltas, decimal.NewFromInt(1))
}

func (s *AccountBalanceRealtimeProjection) applyVoided(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionVoidedPayload](ct)
	if err != nil {
		return err
	}
	entries, err := tx.Projection.TransactionRepo.GetEntriesByTxnId(ctx, p.TransactionId, ct.MerchantID)
	if err != nil {
		return err
	}
	deltas := make([]balanceDelta, len(entries))
	for i, e := range entries {
		deltas[i] = balanceDelta{
			accountId: e.AccountId,
			ledgerId:  e.LedgerId,
			debit:     e.Debit,
			credit:    e.Credit,
		}
	}
	return s.applyDeltas(ctx, tx, ct.MerchantID, deltas, decimal.NewFromInt(-1))
}

func (s *AccountBalanceRealtimeProjection) applyCorrected(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.TransactionCorrectedPayload](ct)
	if err != nil {
		return err
	}
	entries, err := tx.Projection.TransactionRepo.GetEntriesByTxnId(ctx, p.OriginalTransactionId, ct.MerchantID)
	if err != nil {
		return err
	}
	deltas := make([]balanceDelta, len(entries))
	for i, e := range entries {
		deltas[i] = balanceDelta{
			accountId: e.AccountId,
			ledgerId:  e.LedgerId,
			debit:     e.Debit,
			credit:    e.Credit,
		}
	}
	return s.applyDeltas(ctx, tx, ct.MerchantID, deltas, decimal.NewFromInt(-1))
}

func (s *AccountBalanceRealtimeProjection) applyInvestmentBought(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.InvestmentBoughtState](ct)
	if err != nil {
		return err
	}
	return s.applyFromTxnPayload(ctx, tx, ct.MerchantID, &st.Transaction)
}

func (s *AccountBalanceRealtimeProjection) applyInvestmentSold(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.InvestmentSoldState](ct)
	if err != nil {
		return err
	}
	return s.applyFromTxnPayload(ctx, tx, ct.MerchantID, &st.Transaction)
}

func (s *AccountBalanceRealtimeProjection) applyDividendReceived(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	st, err := checkAndGetState[state.DividendReceivedState](ct)
	if err != nil {
		return err
	}
	if st.Transaction == nil {
		return nil
	}
	return s.applyFromTxnPayload(ctx, tx, ct.MerchantID, st.Transaction)
}

func (s *AccountBalanceRealtimeProjection) applyFromTxnPayload(ctx context.Context, tx event_store.EventStoreRepositories, merchantID int64, p *payload.TransactionCreatedPayload) error {
	deltas := make([]balanceDelta, len(p.Entries))
	for i, e := range p.Entries {
		deltas[i] = balanceDelta{
			accountId: e.AccountId,
			ledgerId:  e.LedgerId,
			debit:     e.Debit,
			credit:    e.Credit,
		}
	}
	return s.applyDeltas(ctx, tx, merchantID, deltas, decimal.NewFromInt(1))
}

// applyDeltas applies (sign * delta) to all touched accounts, their ancestors, and ledgers.
func (s *AccountBalanceRealtimeProjection) applyDeltas(
	ctx context.Context,
	tx event_store.EventStoreRepositories,
	merchantID int64,
	deltas []balanceDelta,
	sign decimal.Decimal,
) error {
	accountRepo := tx.Projection.AccountRunningBalanceRepo
	ledgerRepo := tx.Projection.LedgerRunningBalanceRepo

	for _, d := range deltas {
		debit := d.debit.Mul(sign)
		credit := d.credit.Mul(sign)

		// leaf account
		if err := accountRepo.Upsert(ctx, d.accountId, merchantID, debit, credit); err != nil {
			return err
		}

		// ancestor accounts
		ancestors, err := accountRepo.GetAncestorIds(ctx, d.accountId, merchantID)
		if err != nil {
			return err
		}
		for _, ancestorId := range ancestors {
			if err := accountRepo.Upsert(ctx, ancestorId, merchantID, debit, credit); err != nil {
				return err
			}
		}

		// ledger account
		if d.ledgerId != nil {
			if err := ledgerRepo.Upsert(ctx, *d.ledgerId, merchantID, debit, credit); err != nil {
				return err
			}
		}
	}
	return nil
}

type balanceDelta struct {
	accountId string
	ledgerId  *int64
	debit     decimal.Decimal
	credit    decimal.Decimal
}
