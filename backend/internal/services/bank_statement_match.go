package services

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// MatchResult summarises the auto-match run.
type MatchResult struct {
	Matched   int `json:"matched"`
	Fuzzy     int `json:"fuzzy"`
	Unmatched int `json:"unmatched"`
}

type BankStatementMatchService interface {
	// AutoMatch runs the matching algorithm for all UNMATCHED txns in an import.
	AutoMatch(ctx context.Context, importID int64) (*MatchResult, error)
	// ReMatch resets EXACT/FUZZY matches, then re-runs AutoMatch.
	ReMatch(ctx context.Context, importID int64) (*MatchResult, error)
	// ManualMatch assigns a specific entry to a bank txn (MANUAL confidence).
	ManualMatch(ctx context.Context, bankTxnID, entryID int64) error
}

func NewBankStatementMatchService(queryRepo *query.Repo, uow event_store.UnitOfWork) BankStatementMatchService {
	return &bankStatementMatchService{q: queryRepo, uow: uow}
}

type bankStatementMatchService struct {
	q   *query.Repo
	uow event_store.UnitOfWork
}

func (s *bankStatementMatchService) AutoMatch(ctx context.Context, importID int64) (*MatchResult, error) {
	return s.runMatch(ctx, importID)
}

func (s *bankStatementMatchService) ReMatch(ctx context.Context, importID int64) (*MatchResult, error) {
	imp, err := s.q.BankStatementImport.GetBankStatementImportByID(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("get import: %w", err)
	}
	if imp == nil {
		return nil, fmt.Errorf("import not found")
	}
	// Reset EXACT/FUZZY (not CONFIRMED/MANUAL) before re-matching
	if err := s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
		return tx.Projection.BankStatementImportRepo.ResetNonConfirmedMatches(ctx, sqlcdb.ResetNonConfirmedMatchesParams{
			ImportID:   imp.ImportID,
			MerchantID: imp.MerchantID,
		})
	}); err != nil {
		return nil, fmt.Errorf("reset matches: %w", err)
	}
	return s.runMatch(ctx, importID)
}

func (s *bankStatementMatchService) ManualMatch(ctx context.Context, bankTxnID, entryID int64) error {
	txn, err := s.q.BankStatementImport.GetBankStatementTxnByID(ctx, bankTxnID)
	if err != nil {
		return fmt.Errorf("get bank txn: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("bank statement txn not found")
	}
	confidence := "MANUAL"
	return s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
		return tx.Projection.BankStatementImportRepo.UpdateBankTxnMatch(ctx, sqlcdb.UpdateBankTxnMatchParams{
			BankTxnID:       bankTxnID,
			MerchantID:      txn.MerchantID,
			MatchStatus:     "MATCHED",
			MatchedEntryID:  &entryID,
			MatchConfidence: &confidence,
		})
	})
}

func (s *bankStatementMatchService) runMatch(ctx context.Context, importID int64) (*MatchResult, error) {
	imp, err := s.q.BankStatementImport.GetBankStatementImportByID(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("get import: %w", err)
	}
	if imp == nil {
		return nil, fmt.Errorf("import not found")
	}

	bankTxns, err := s.q.BankStatementImport.GetUnmatchedBankTxns(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("get unmatched txns: %w", err)
	}
	if len(bankTxns) == 0 {
		return &MatchResult{}, nil
	}

	dateFrom, dateTo := txnDateRange(bankTxns, imp.StatementDate)
	entries, err := s.q.BankStatementImport.GetLedgerEntriesForMatching(ctx, imp.LedgerID, dateFrom, dateTo)
	if err != nil {
		return nil, fmt.Errorf("get ledger entries: %w", err)
	}

	byDate := make(map[string][]sqlcdb.GetLedgerEntriesForMatchingRow)
	for _, e := range entries {
		byDate[e.TxnDate] = append(byDate[e.TxnDate], e)
	}

	usedEntryIDs := make(map[int64]bool)
	result := &MatchResult{}
	type pending struct {
		params     sqlcdb.UpdateBankTxnMatchParams
		merchantID int64
	}
	var updates []sqlcdb.UpdateBankTxnMatchParams

	for _, bt := range bankTxns {
		btAmt := bankTxnAmount(bt)

		// Step 1: EXACT match
		candidates := filterEntries(byDate[bt.TxnDate], btAmt, usedEntryIDs)
		if len(candidates) == 1 {
			usedEntryIDs[candidates[0].EntryID] = true
			conf := "EXACT"
			updates = append(updates, sqlcdb.UpdateBankTxnMatchParams{
				BankTxnID:       bt.BankTxnID,
				MerchantID:      bt.MerchantID,
				MatchStatus:     "MATCHED",
				MatchedEntryID:  &candidates[0].EntryID,
				MatchConfidence: &conf,
			})
			result.Matched++
			continue
		}

		// Step 2: FUZZY match (±3 days)
		fuzzy := fuzzyEntries(byDate, bt.TxnDate, btAmt, usedEntryIDs, 3)
		if len(fuzzy) == 1 {
			usedEntryIDs[fuzzy[0].EntryID] = true
			conf := "FUZZY"
			updates = append(updates, sqlcdb.UpdateBankTxnMatchParams{
				BankTxnID:       bt.BankTxnID,
				MerchantID:      bt.MerchantID,
				MatchStatus:     "MATCHED",
				MatchedEntryID:  &fuzzy[0].EntryID,
				MatchConfidence: &conf,
			})
			result.Fuzzy++
			continue
		}

		result.Unmatched++
	}

	// Batch-write match results in one UoW transaction
	if len(updates) > 0 {
		if err := s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
			for _, u := range updates {
				if err := tx.Projection.BankStatementImportRepo.UpdateBankTxnMatch(ctx, u); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("persist match results: %w", err)
		}
	}
	return result, nil
}

// ─────────────────────────────────────────
// helpers
// ─────────────────────────────────────────

func bankTxnAmount(bt projection.BankStatementTxn) decimal.Decimal {
	if bt.Credit.IsPositive() {
		return bt.Credit
	}
	return bt.Debit
}

func filterEntries(entries []sqlcdb.GetLedgerEntriesForMatchingRow, amt decimal.Decimal, used map[int64]bool) []sqlcdb.GetLedgerEntriesForMatchingRow {
	var result []sqlcdb.GetLedgerEntriesForMatchingRow
	for _, e := range entries {
		if used[e.EntryID] {
			continue
		}
		entryAmt := e.Debit
		if e.Credit.IsPositive() {
			entryAmt = e.Credit
		}
		if entryAmt.Equal(amt) {
			result = append(result, e)
		}
	}
	return result
}

func fuzzyEntries(byDate map[string][]sqlcdb.GetLedgerEntriesForMatchingRow, baseDate string, amt decimal.Decimal, used map[int64]bool, dayRange int) []sqlcdb.GetLedgerEntriesForMatchingRow {
	base, err := time.Parse("2006-01-02", baseDate)
	if err != nil {
		return nil
	}
	var result []sqlcdb.GetLedgerEntriesForMatchingRow
	for d := -dayRange; d <= dayRange; d++ {
		if d == 0 {
			continue
		}
		checkDate := base.AddDate(0, 0, d).Format("2006-01-02")
		result = append(result, filterEntries(byDate[checkDate], amt, used)...)
	}
	return result
}

func txnDateRange(txns []projection.BankStatementTxn, statementDate string) (string, string) {
	if len(txns) == 0 {
		return statementDate, statementDate
	}
	from, to := txns[0].TxnDate, txns[0].TxnDate
	for _, t := range txns[1:] {
		if t.TxnDate < from {
			from = t.TxnDate
		}
		if t.TxnDate > to {
			to = t.TxnDate
		}
	}
	if f, err := time.Parse("2006-01-02", from); err == nil {
		from = f.AddDate(0, 0, -3).Format("2006-01-02")
	}
	if t, err := time.Parse("2006-01-02", to); err == nil {
		to = t.AddDate(0, 0, 3).Format("2006-01-02")
	}
	return from, to
}
