package event_store

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxTxProjectionRepository struct{ tx *sqlx.Tx }

func (r *sqlxTxProjectionRepository) UpsertTransaction(ctx context.Context, p projection.Transaction) error {
	_, err := r.tx.NamedExecContext(ctx, `
        INSERT INTO proj_transactions
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
            INSERT INTO proj_journal_entries
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
            INSERT INTO proj_journal_entries
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

func (r *sqlxTxProjectionRepository) CreateAccount(ctx context.Context, p projection.Account) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO proj_accounts
                (account_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, note, version)
            VALUES (:account_id, :parent_id, :name, :type, :normal_balance, :currency, :is_summary, :is_active, :note, 1)`,
		p,
	)
	return err
}

func (r *sqlxTxProjectionRepository) UpdateAccount(ctx context.Context, p projection.Account) error {
	_, err := r.tx.NamedExecContext(ctx, `
            UPDATE proj_ledger_accounts
                SET parent_id = :parent_id, 
                    name = :name, 
                    type = :type, 
                    normal_balance = :normal_balance, 
                    currency = :currency, 
                    is_summary = :is_summary, 
                    is_active = :is_active, 
                    note = :note, 
                    version = varsion + 1
            WHERE account_id = :account_id
            AND version = :version`,
		p,
	)
	return err
}

func (r *sqlxTxProjectionRepository) CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO proj_ledger_accounts
                (account_id, institution, name, account_no, currency, credit_limit, billing_day, due_day, is_active, note, version)
            VALUES (:account_id, :institution, :name, :account_no, :currency, :credit_limit, :billing_day, :due_day, :is_active, :note, 1)`,
		p,
	)
	return err
}

func (r *sqlxTxProjectionRepository) UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	_, err := r.tx.NamedExecContext(ctx, `
            UPDATE proj_ledger_accounts
                SET account_id = :account_id, 
                    institution = :institution, 
                    name = :name, 
                    account_no = :account_no, 
                    currency = :currency, 
                    credit_limit = :credit_limit, 
                    billing_day = :billing_day, 
                    due_day = :due_day, 
                    is_active = :is_active, 
                    note = :note, 
                    version = varsion + 1
            WHERE ledger_id = :ledger_id
            AND version = :version`,
		p,
	)
	return err
}
