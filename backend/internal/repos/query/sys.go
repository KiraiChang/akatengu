package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/ctxkey"
	"context"

	"github.com/jmoiron/sqlx"
)

type SysRepo interface {
	GetSysAccount(ctx context.Context) ([]db.SysAccount, error)
}

type sqlcdbSysRepo struct {
	q *sqlcdb.Queries
}

func newSysRepo(q *sqlcdb.Queries) SysRepo {
	return &sqlcdbSysRepo{q: q}
}

func NewSysRepo(db *sqlx.DB) SysRepo {
	return &sqlcdbSysRepo{q: sqlcdb.New(db)}
}

func (s *sqlcdbSysRepo) GetSysAccount(ctx context.Context) ([]db.SysAccount, error) {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.q.GetSysAccounts(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	result := make([]db.SysAccount, len(rows))
	for i, row := range rows {
		result[i] = db.SysAccountFromGetSysAccountsRow(row)
	}
	return result, nil
}
