package projection_repo

import (
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxAccountRepo struct {
	tx *sqlx.Tx
}

func NewAccountRepo(tx *sqlx.Tx) AccountRepo {
	return &sqlxAccountRepo{
		tx: tx,
	}
}

func (r *sqlxAccountRepo) CreateAccount(ctx context.Context, p projection.Account) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO accounts
                (account_id, parent_id, name, type, normal_balance, currency, is_summary, is_active, note, version)
            VALUES (:account_id, :parent_id, :name, :type, :normal_balance, :currency, :is_summary, :is_active, :note, 1)`,
		p,
	)
	return err
}

func (r *sqlxAccountRepo) UpdateAccount(ctx context.Context, p projection.Account) error {
	_, err := r.tx.NamedExecContext(ctx, `
            UPDATE ledger_accounts
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

func (r *sqlxAccountRepo) CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	_, err := r.tx.NamedExecContext(ctx, `
            INSERT INTO ledger_accounts
                (account_id, institution, name, account_no, currency, credit_limit, billing_day, due_day, is_active, note, version)
            VALUES (:account_id, :institution, :name, :account_no, :currency, :credit_limit, :billing_day, :due_day, :is_active, :note, 1)`,
		p,
	)
	return err
}

func (r *sqlxAccountRepo) UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error {
	_, err := r.tx.NamedExecContext(ctx, `
            UPDATE ledger_accounts
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
