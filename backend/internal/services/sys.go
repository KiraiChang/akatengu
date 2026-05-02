package services

import (
	"akatengu/internal/model/db"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type SysService interface {
	GetSysAccount(ctx context.Context) ([]db.SysAccount, error)
	UpdateSysAccount(ctx context.Context, s db.SysAccount) error
}

func NewSysService(db *sqlx.DB) SysService {
	return &sysService{
		q: query.NewSysRepo(db),
	}
}

type sysService struct {
	q query.SysRepo
}

func (s sysService) GetSysAccount(ctx context.Context) ([]db.SysAccount, error) {
	return s.q.GetSysAccount(ctx)
}

func (s sysService) UpdateSysAccount(ctx context.Context, sys db.SysAccount) error {
	return s.q.UpdateSysAccount(ctx, sys)
}
