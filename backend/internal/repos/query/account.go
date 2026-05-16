package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/ctxkey"
	"context"

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
}

type sqlxAccountRepo struct {
	q  *sqlcdb.Queries
	db *sqlx.DB
}

func (r *sqlxAccountRepo) GetLedgerPaged(ctx context.Context, params model.PaginationParams) ([]projection.LedgerAccount, int64, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.GetLedgerPaged(ctx, sqlcdb.GetLedgerPagedParams{
		MerchantID: merchantID,
		Limit:      params.Limit,
		Offset:     params.Offset,
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.q.GetChildrenAccount(ctx, sqlcdb.GetChildrenAccountParams{
		ParentID:   &id,
		MerchantID: merchantID,
	})
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.q.GetAccountPaged(ctx, sqlcdb.GetAccountPagedParams{
		MerchantID: merchantID,
		Limit:      params.Limit,
		Offset:     params.Offset,
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

func (r *sqlxAccountRepo) GetAllLedgers(ctx context.Context) ([]projection.LedgerAccount, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllLedgers(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]projection.LedgerAccount, len(rows))
	for i, row := range rows {
		result[i] = projection.LedgerAccountFromGetAllLedgersRow(row)
	}
	return result, nil
}

func (r *sqlxAccountRepo) GetAllAccounts(ctx context.Context) ([]projection.Account, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.GetAllAccounts(ctx, merchantID)
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
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetAccount(ctx, sqlcdb.GetAccountParams{
		AccountID:  id,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	a := projection.AccountFromGetAccountRow(row)
	return &a, nil
}

func (r *sqlxAccountRepo) GetLedger(ctx context.Context, id int64) (*projection.LedgerAccount, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetLedger(ctx, sqlcdb.GetLedgerParams{
		LedgerID:   id,
		MerchantID: merchantID,
	})
	if err != nil {
		return nil, err
	}
	la := projection.LedgerAccountFromGetLedgerRow(row)
	return &la, nil
}
