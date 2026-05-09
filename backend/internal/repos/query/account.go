package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

type AccountRepo interface {
	GetAccount(ctx context.Context, id string) (*projection.Account, error)
	GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error)
	GetAllAccounts(ctx context.Context) ([]projection.Account, error)
	GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error)
	GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error)
	GetLedgerPaged(ctx context.Context, req model.PaginationParams) ([]projection.LedgerAccount, int64, error)
	GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error)
	GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error)
	GetAccountSummaryOptimized(ctx context.Context) ([]projection.AccountBalance, error)
}

type sqlxAccountRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func (r *sqlxAccountRepo) GetAllLedgerBalances(ctx context.Context) ([]projection.LedgerAccountBalance, error) {
	rows, err := r.q.GetAllLedgerBalances(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccountBalance, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountBalanceFromGetAllLedgerBalancesRow(row)
		if result[i].NormalBalance.Is(enums.BalanceDebit) {
			result[i].Balance = result[i].DebitTotal.Sub(result[i].CreditTotal)
		} else {
			result[i].Balance = result[i].CreditTotal.Sub(result[i].DebitTotal)
		}
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetLedgerPaged(ctx context.Context, params model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	rows, err := r.q.GetLedgerPaged(ctx, sqlcdb.GetLedgerPagedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.LedgerAccount, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountFromGetLedgerPagedRow(row)
	}
	return result, total, nil
}

func (r *sqlxAccountRepo) GetChildrenAccount(ctx context.Context, id string) ([]projection.Account, error) {
	rows, err := r.q.GetChildrenAccount(ctx, &id)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetChildrenAccountRow(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAccountPaged(ctx context.Context, params model.PaginationParams) ([]projection.Account, int64, error) {
	rows, err := r.q.GetAccountPaged(ctx, sqlcdb.GetAccountPagedParams{
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	total := int64(0)
	if err != nil {
		return nil, total, err
	}
	if len(rows) > 0 {
		total = rows[0].Total
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetAccountPagedRow(row)
	}
	return result, total, nil
}

func newAccountRepo(q *sqlcdb.Queries, db *sqlx.DB) AccountRepo {
	return &sqlxAccountRepo{q: q, db: db}
}

func NewAccountRepo(db *sqlx.DB) AccountRepo {
	return &sqlxAccountRepo{q: sqlcdb.New(db), db: db}
}

const queryAccountSummaryFull = `
SELECT a.account_id, a.name, a.type, a.normal_balance,
       COALESCE(SUM(je.debit), 0)  AS debit_total,
       COALESCE(SUM(je.credit), 0) AS credit_total
FROM accounts a
    LEFT JOIN journal_entries je ON a.account_id = je.account_id
    LEFT JOIN transactions t     ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
WHERE a.is_active = 1
GROUP BY a.account_id`

const queryAccountSummaryWithSnapshot = `
WITH snap AS (
    SELECT account_id, debit_total, credit_total
    FROM account_balance_snapshots
    WHERE closing_id = ?
),
delta AS (
    SELECT je.account_id,
           COALESCE(SUM(je.debit), 0)  AS delta_debit,
           COALESCE(SUM(je.credit), 0) AS delta_credit
    FROM journal_entries je
    JOIN transactions t ON je.txn_id = t.txn_id AND t.status = 'ACTIVE'
    WHERE t.txn_date > ?
    GROUP BY je.account_id
)
SELECT
    a.account_id, a.name, a.type, a.normal_balance,
    COALESCE(snap.debit_total, 0)   + COALESCE(delta.delta_debit, 0)   AS debit_total,
    COALESCE(snap.credit_total, 0)  + COALESCE(delta.delta_credit, 0)  AS credit_total
FROM accounts a
    LEFT JOIN snap  ON a.account_id = snap.account_id
    LEFT JOIN delta ON a.account_id = delta.account_id
WHERE a.is_active = 1`

func (r *sqlxAccountRepo) GetAccountSummaryOptimized(ctx context.Context) ([]projection.AccountBalance, error) {
	var latest struct {
		ClosingId int64  `db:"closing_id"`
		PeriodEnd string `db:"period_end"`
	}
	err := r.db.GetContext(ctx, &latest, `
		SELECT closing_id, period_end FROM period_closings
		WHERE status = 'CLOSED' AND period_type = 'MONTHLY'
		ORDER BY period_end DESC LIMIT 1`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var rows []projection.AccountBalance
			return rows, r.db.SelectContext(ctx, &rows, queryAccountSummaryFull)
		}
		return nil, err
	}
	var rows []projection.AccountBalance
	return rows, r.db.SelectContext(ctx, &rows, queryAccountSummaryWithSnapshot, latest.ClosingId, latest.PeriodEnd)
}

func (r *sqlxAccountRepo) GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error) {
	rows, err := r.q.GetAllLedgers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccount, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountFromLedgerAccount(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAllAccounts(ctx context.Context) ([]projection.Account, error) {
	rows, err := r.q.GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]projection.Account, len(rows))
	for i, row := range rows {
		result[i] = projection.AccountFromGetAllAccountsRow(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAccount(ctx context.Context, id string) (*projection.Account, error) {
	row, err := r.q.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	a := projection.AccountFromAccount(row)
	return &a, nil
}

func (r *sqlxAccountRepo) GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error) {
	row, err := r.q.GetLedger(ctx, id)
	if err != nil {
		return nil, err
	}
	la := projection.LedgerAccountFromLedgerAccount(row)
	return &la, nil
}
