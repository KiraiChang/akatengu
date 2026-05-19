package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
	"fmt"
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
	for _, e := range entries {
		if e.EntryId != 0 {
			err := r.q.InsertJournalEntryWithID(ctx, e.ToInsertJournalEntryWithIDParams())
			if err != nil {
				return err
			}
		} else {
			err := r.q.InsertJournalEntry(ctx, e.ToInsertJournalEntryParams())
			if err != nil {
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
