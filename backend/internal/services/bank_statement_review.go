package services

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ReviewItem is one unmatched bank txn with a suggested journal entry preview.
type ReviewItem struct {
	BankTxnID    int64           `json:"bank_txn_id"`
	BankTxnUUID  string          `json:"bank_txn_uuid"`
	TxnDate      string          `json:"txn_date"`
	Description  string          `json:"description"`
	Debit        decimal.Decimal `json:"debit"`
	Credit       decimal.Decimal `json:"credit"`
	MatchStatus  string          `json:"match_status"`
	SuggestedEntry *SuggestedEntry `json:"suggested_entry"`
}

// SuggestedEntry is a preview of the journal entry to be created on approval.
type SuggestedEntry struct {
	LedgerAccountID string          `json:"ledger_account_id"` // the bank account's account_id
	LedgerID        int64           `json:"ledger_id"`
	CounterAccountID string         `json:"counter_account_id"` // suggested opposite account
	Amount          decimal.Decimal `json:"amount"`
	IsDebit         bool            `json:"is_debit"` // true if bank txn is a withdrawal
}

// ApproveRequest contains user-selected accounts for the journal entry.
type ApproveRequest struct {
	BankTxnID       int64   `json:"bank_txn_id"`
	LedgerID        int64   `json:"ledger_id"`
	AccountID       string  `json:"account_id"`       // ledger's account_id (debit or credit side)
	CounterAccountID string `json:"counter_account_id"` // the other side
	TxnDate         string  `json:"txn_date"`
	Description     string  `json:"description"`
	Note            *string `json:"note"`
}

type BankStatementReviewService interface {
	GetReviewItems(ctx context.Context, importID int64) ([]ReviewItem, error)
	ApproveTxn(ctx context.Context, req ApproveRequest) error
	IgnoreTxn(ctx context.Context, bankTxnID int64) error
}

func NewBankStatementReviewService(queryRepo *query.Repo, uow event_store.UnitOfWork, es *EventStoreService) BankStatementReviewService {
	return &bankStatementReviewService{q: queryRepo, uow: uow, es: es}
}

type bankStatementReviewService struct {
	q   *query.Repo
	uow event_store.UnitOfWork
	es  *EventStoreService
}

func (s *bankStatementReviewService) GetReviewItems(ctx context.Context, importID int64) ([]ReviewItem, error) {
	imp, err := s.q.BankStatementImport.GetBankStatementImportByID(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("get import: %w", err)
	}
	if imp == nil {
		return nil, fmt.Errorf("import not found")
	}

	// Get ledger account info for suggestion
	ledger, err := s.q.Account.GetLedger(ctx, imp.LedgerID)
	if err != nil {
		return nil, fmt.Errorf("get ledger: %w", err)
	}

	txns, err := s.q.BankStatementImport.GetBankTxnsForReview(ctx, importID)
	if err != nil {
		return nil, fmt.Errorf("get review txns: %w", err)
	}

	items := make([]ReviewItem, 0, len(txns))
	for _, txn := range txns {
		item := ReviewItem{
			BankTxnID:   txn.BankTxnID,
			BankTxnUUID: txn.BankTxnUUID,
			TxnDate:     txn.TxnDate,
			Description: txn.Description,
			Debit:       txn.Debit,
			Credit:      txn.Credit,
			MatchStatus: txn.MatchStatus,
		}
		if txn.MatchStatus == "UNMATCHED" && ledger != nil {
			amt := txn.Debit
			isDebit := true
			if txn.Credit.IsPositive() {
				amt = txn.Credit
				isDebit = false
			}
			item.SuggestedEntry = &SuggestedEntry{
				LedgerAccountID: ledger.AccountId,
				LedgerID:        imp.LedgerID,
				CounterAccountID: suggestCounterAccount(txn.Description),
				Amount:          amt,
				IsDebit:         isDebit,
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *bankStatementReviewService) ApproveTxn(ctx context.Context, req ApproveRequest) error {
	txn, err := s.q.BankStatementImport.GetBankStatementTxnByID(ctx, req.BankTxnID)
	if err != nil {
		return fmt.Errorf("get bank txn: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("bank statement txn not found")
	}
	if txn.MatchStatus == "APPROVED" {
		return fmt.Errorf("bank txn is already approved")
	}
	if txn.MatchStatus == "IGNORED" {
		return fmt.Errorf("bank txn is already ignored")
	}

	// Determine debit/credit sides based on whether bank txn is withdrawal or deposit
	var entries []payload.TransactionEntryPayload
	amt := txn.Debit
	if txn.Credit.IsPositive() {
		amt = txn.Credit
		// Deposit: bank account debits, counter account credits
		entries = []payload.TransactionEntryPayload{
			{AccountId: req.AccountID, LedgerId: &req.LedgerID, Debit: amt, Credit: decimal.Zero},
			{AccountId: req.CounterAccountID, Debit: decimal.Zero, Credit: amt},
		}
	} else {
		// Withdrawal: counter account debits, bank account credits
		entries = []payload.TransactionEntryPayload{
			{AccountId: req.CounterAccountID, Debit: amt, Credit: decimal.Zero},
			{AccountId: req.AccountID, LedgerId: &req.LedgerID, Debit: decimal.Zero, Credit: amt},
		}
	}

	txnPayload := payload.TransactionCreatedPayload{
		TransactionDate: req.TxnDate,
		Description:     req.Description,
		TotalAmount:     amt,
		Currency:        "TWD",
		Note:            req.Note,
		Entries:         entries,
	}
	if err := txnPayload.Validate(); err != nil {
		return fmt.Errorf("transaction payload invalid: %w", err)
	}

	b, err := json.Marshal(txnPayload)
	if err != nil {
		return fmt.Errorf("marshal transaction payload: %w", err)
	}

	txnAggID, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate transaction uuid: %w", err)
	}

	// Append transaction.created event
	_, err = s.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateTransaction.Enum(),
		AggregateID:     txnAggID.String(),
		ExpectedVersion: 0,
		EventType:       event_types.EventTransactionCreated.Enum(),
		Payload:         b,
	})
	if err != nil {
		return fmt.Errorf("create transaction: %w", err)
	}

	// Lookup the created transaction by its aggregate UUID (txnAggID == txn_uuid in transactions table)
	createdTxn, err := s.q.Transaction.GetByUUID(ctx, txnAggID.String())
	if err != nil {
		return fmt.Errorf("get created transaction: %w", err)
	}
	if createdTxn == nil {
		return fmt.Errorf("created transaction not found after event")
	}

	// Mark bank txn as APPROVED with created_txn_id
	return s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
		createdTxnID := createdTxn.TransactionId
		return tx.Projection.BankStatementImportRepo.UpdateBankTxnCreatedTxn(ctx, sqlcdb.UpdateBankTxnCreatedTxnParams{
			MerchantID:   txn.MerchantID,
			BankTxnID:    req.BankTxnID,
			CreatedTxnID: &createdTxnID,
		})
	})
}

func (s *bankStatementReviewService) IgnoreTxn(ctx context.Context, bankTxnID int64) error {
	txn, err := s.q.BankStatementImport.GetBankStatementTxnByID(ctx, bankTxnID)
	if err != nil {
		return fmt.Errorf("get bank txn: %w", err)
	}
	if txn == nil {
		return fmt.Errorf("bank statement txn not found")
	}
	return s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
		return tx.Projection.BankStatementImportRepo.UpdateBankTxnStatus(ctx, sqlcdb.UpdateBankTxnStatusParams{
			MerchantID:  txn.MerchantID,
			BankTxnID:   bankTxnID,
			MatchStatus: "IGNORED",
		})
	})
}

// suggestCounterAccount suggests an account_id based on description keywords.
func suggestCounterAccount(description string) string {
	// TODO: implement keyword-based suggestion from system account settings
	// For now return empty string; frontend should prompt user to select
	return ""
}
