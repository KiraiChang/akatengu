package projection

import (
	"akatengu/internal/model/db/projection"
	"context"
)

func (r *sqlxTxProjectionRepository) UpsertTransaction(ctx context.Context, p projection.Transaction) error {
	_, err := r.tx.NamedExecContext(ctx, `
        INSERT INTO transactions
            (txn_id, txn_date, description, total_amount, currency, status, receipt_no, note, version)
        VALUES (:txn_id, :txn_date, :description, :total_amount, :currency, :status, :receipt_no, :note, :version)
        ON CONFLICT(txn_id) DO UPDATE SET
            txn_date     = excluded.txn_date,
            description  = excluded.description,
            total_amount = excluded.total_amount,
            status       = excluded.status,
            version      = excluded.version`,
		p,
	)
	return err
}

func (r *sqlxTxProjectionRepository) UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error {
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
		}
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
	return nil
}
