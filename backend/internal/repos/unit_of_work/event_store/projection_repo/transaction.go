package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"
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

func (r *sqlxTransactionRepo) UpdateTxnStatus(ctx context.Context, txnId int64, status enums.TransactionStatus, version int64) error {
	return r.q.UpdateTransactionStatus(ctx, sqlcdb.UpdateTransactionStatusParams{
		Status:  status,
		Version: version,
		TxnID:   txnId,
	})
}

func (r *sqlxTransactionRepo) SysUpdateTxnStatus(ctx context.Context, txnId int64, refTxtId *int64, status enums.TransactionStatus) error {
	return r.q.SysUpdateTransactionStatus(ctx, sqlcdb.SysUpdateTransactionStatusParams{
		Status:   status,
		RefTxnID: refTxtId,
		TxnID:    txnId,
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
