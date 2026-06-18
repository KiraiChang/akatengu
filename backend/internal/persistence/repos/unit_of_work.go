package repos

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/persistence/repos/projection"
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type UnitOfWork interface {
	// Do 開啟 transaction，執行 fn，自動 commit 或 rollback
	Do(ctx context.Context, fn func(tx *Transaction) error) error
}

type sqlxUnitOfWork struct {
	db *sqlx.DB
}

func NewUnitOfWork(db *sqlx.DB) UnitOfWork {
	return &sqlxUnitOfWork{db: db}
}

func (u *sqlxUnitOfWork) Do(ctx context.Context, fn func(*Transaction) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	q := sqlcdb.New(tx)
	repos := &Transaction{
		Store:      &sqlxStore{q: q},
		Snap:       &sqlcdbTxSnapshot{q: q},
		Check:      &sqlcdbTxCheckpoint{q: q},
		Projection: projection.NewProjection(q),
	}

	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit()
}
