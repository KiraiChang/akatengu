package event_store

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"akatengu/internal/database/sqlcdb"
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

	q := sqlcdb.New(tx)
	repos := EventStoreRepositories{
		Event:      &sqlcdbTxEventRepository{q: q},
		Version:    &sqlcdbTxVersionRepository{q: q},
		Snap:       &sqlcdbTxSnapshotRepository{q: q},
		Check:      &sqlcdbTxCheckpointRepository{q: q},
		Projection: projection_repo.NewTxProjectionRepository(q),
		Truncate:   &sqlxTruncateRepository{tx: tx},
	}

	if err := fn(repos); err != nil {
		return err
	}

	return tx.Commit()
}
