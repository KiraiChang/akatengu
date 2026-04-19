package repos

import (
	"akatengu/internal/model/db"
	"context"

	"github.com/jmoiron/sqlx"
)

type SysRepo interface {
	GetSysAccount(ctx context.Context) ([]db.SysAccount, error)
}

type sqlSysRepo struct {
	db *sqlx.DB
}

func (s sqlSysRepo) GetSysAccount(ctx context.Context) ([]db.SysAccount, error) {
	var result []db.SysAccount
	err := s.db.SelectContext(ctx, &result, `
        SELECT *
        FROM sys_accounts`,
	)
	return result, err
}

func NewSysRepo(db *sqlx.DB) SysRepo {
	return &sqlSysRepo{db: db}
}
