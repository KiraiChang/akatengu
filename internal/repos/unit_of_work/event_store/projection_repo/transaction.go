package projection_repo

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxTransactionRepo struct {
	tx *sqlx.Tx
}

func NewTransactionRepo(tx *sqlx.Tx) TransactionRepo {
	return &sqlxTransactionRepo{tx: tx}
}

func (r *sqlxTransactionRepo) InsertTxn(ctx context.Context, p projection.Transaction) (int64, error) {
	result, err := r.tx.NamedExecContext(ctx, `
        INSERT INTO transactions
            (txn_date, description, total_amount, currency, status, receipt_no, note, version, ref_txn_id)
        VALUES (:txn_date, :description, :total_amount, :currency, :status, :receipt_no, :note, 1, :ref_txn_id)`,
		p,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *sqlxTransactionRepo) UpdateTxnStatus(ctx context.Context, txnId int64, status enums.TransactionStatus, version int64) error {
	_, err := r.tx.ExecContext(ctx, `
	UPDATE transactions
		SET status = ?,
		    version = version + 1
		WHERE txn_id = ?
			AND version = ?`,
		status,
		txnId,
		version)
	return err
}

func (r *sqlxTransactionRepo) SysUpdateTxnStatus(ctx context.Context, txnId int64, refTxtId *int64, status enums.TransactionStatus) error {
	_, err := r.tx.ExecContext(ctx, `
	UPDATE transactions
		SET status = ?,
			ref_txn_id = ?,
		    version = version + 1
		WHERE txn_id = ?`,
		status,
		refTxtId,
		txnId)

	return err
}

func (r *sqlxTransactionRepo) UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error {
	for _, e := range entries {
		if e.EntryId != 0 {
			_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO journal_entries
                (entry_id, txn_id, ledger_id, account_id, debit, credit, note)
            VALUES (:entry_id, :txn_id, :ledger_id, :account_id, :debit, :credit, :note)
            ON CONFLICT(entry_id) DO NOTHING`,
				e,
			)
			if err != nil {
				return err
			}
		} else {
			_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO journal_entries
                (txn_id, ledger_id, account_id, debit, credit, note)
            VALUES (:txn_id, :ledger_id, :account_id, :debit, :credit, :note)`,
				e,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
