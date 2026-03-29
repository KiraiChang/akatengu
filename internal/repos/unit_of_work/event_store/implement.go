package event_store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/repos/unit_of_work/event_store/projection_repo"
)

type sqlxUnitOfWork struct {
	db *sqlx.DB
}

func NewUnitOfWork(db *sqlx.DB) UnitOfWork {
	return &sqlxUnitOfWork{db: db}
}

func (u *sqlxUnitOfWork) Do(ctx context.Context, fn func(EventStoreRepositories) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 把 tx 包進各個 repository，傳給 fn
	repos := EventStoreRepositories{
		Event:      &sqlxTxEventRepository{tx: tx},
		Version:    &sqlxTxVersionRepository{tx: tx},
		Snap:       &sqlxTxSnapshotRepository{tx: tx},
		Check:      &sqlxTxCheckpointRepository{tx: tx},
		Projection: projection_repo.NewTxProjectionRepository(tx),
	}

	if err := fn(repos); err != nil {
		return err // defer 自動 rollback
	}

	return tx.Commit()
}
