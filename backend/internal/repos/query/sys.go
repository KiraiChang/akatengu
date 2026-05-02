package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/model/db"
	"context"

	"github.com/jmoiron/sqlx"
)

type SysRepo interface {
	GetSysAccount(ctx context.Context) ([]db.SysAccount, error)
	UpdateSysAccount(ctx context.Context, s db.SysAccount) error
}

type sqlcdbSysRepo struct {
	q *sqlcdb.Queries
}

func (s *sqlcdbSysRepo) UpdateSysAccount(ctx context.Context, sys db.SysAccount) error {
	return s.q.UpdateSysAccount(ctx, sys.ToUpdateSysAccountParams())
}

func newSysRepo(q *sqlcdb.Queries) SysRepo {
	return &sqlcdbSysRepo{q: q}
}

func NewSysRepo(db *sqlx.DB) SysRepo {
	return &sqlcdbSysRepo{q: sqlcdb.New(db)}
}

func (s *sqlcdbSysRepo) GetSysAccount(ctx context.Context) ([]db.SysAccount, error) {
	rows, err := s.q.GetSysAccounts(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]db.SysAccount, len(rows))
	for i, row := range rows {
		result[i] = db.SysAccountFromSysAccount(row)
	}
	return result, nil
}
