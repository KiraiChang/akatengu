package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"fmt"
	"log"
)

type sqlxTransactionRepo struct {
	q *sqlcdb.Queries
}

func NewTransactionRepo(q *sqlcdb.Queries) TransactionRepo {
	return &sqlxTransactionRepo{q}
}

func (r *sqlxTransactionRepo) InsertTxn(ctx context.Context, p projection.Transaction) (int64, error) {
	return r.q.InsertTransaction(ctx, p.ToInsertTransactionParams())
}

func (r *sqlxTransactionRepo) UpdateTxnStatus(ctx context.Context, txnId int64, status enums.TransactionStatus, version int64, updatedBy *string) error {
	return r.q.UpdateTransactionStatus(ctx, sqlcdb.UpdateTransactionStatusParams{
		Status:    status,
		UpdatedBy: updatedBy,
		Version:   version,
		TxnID:     txnId,
	})
}

func (r *sqlxTransactionRepo) SysUpdateTxnStatus(ctx context.Context, txnId int64, refTxtId *int64, status enums.TransactionStatus, updatedBy *string) error {
	return r.q.SysUpdateTransactionStatus(ctx, sqlcdb.SysUpdateTransactionStatusParams{
		Status:    status,
		RefTxnID:  refTxtId,
		UpdatedBy: updatedBy,
		TxnID:     txnId,
	})
}

func (r *sqlxTransactionRepo) UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error {
	for i := range entries {
		if entries[i].LedgerId != nil && entries[i].LedgerUuid == "" {
			ledger, err := r.q.GetLedger(ctx, sqlcdb.GetLedgerParams{
				LedgerID:   *entries[i].LedgerId,
				MerchantID: entries[i].MerchantID,
			})
			if err == nil {
				entries[i].LedgerUuid = ledger.LedgerUuid
			} else {
				log.Printf("warn: UpsertJournalEntries lookup ledger_uuid for ledger_id=%d: %v", *entries[i].LedgerId, err)
			}
		}
		if entries[i].EntryId != 0 {
			if err := r.q.InsertJournalEntryWithID(ctx, entries[i].ToInsertJournalEntryWithIDParams()); err != nil {
				return err
			}
		} else {
			if err := r.q.InsertJournalEntry(ctx, entries[i].ToInsertJournalEntryParams()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *sqlxTransactionRepo) GetEntriesByTxnId(ctx context.Context, txnId int64, merchantID int64) ([]projection.Entry, error) {
	rows, err := r.q.GetJournalEntries(ctx, sqlcdb.GetJournalEntriesParams{
		TxnID:      txnId,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetEntriesByTxnId: %w", err)
	}
	entries := make([]projection.Entry, 0, len(rows))
	for _, row := range rows {
		e := projection.EntryFromGetJournalEntriesRow(row)
		e.MerchantID = merchantID
		entries = append(entries, e)
	}
	return entries, nil
}
