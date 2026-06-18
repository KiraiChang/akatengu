package eventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"

	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/kernel/event"
	"akatengu/internal/pkg/ctxkey"
)

// Store appends a single event to the event store with optimistic locking.
// Each Append call opens its own transaction.
type Store interface {
	Append(ctx context.Context, evt event.Event) error
}

type sqlxStore struct {
	db *sqlx.DB
}

func NewStore(db *sqlx.DB) Store {
	return &sqlxStore{db: db}
}

func (s *sqlxStore) Append(ctx context.Context, evt event.Event) error {
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return fmt.Errorf("eventstore: %w", err)
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("eventstore: begin tx: %w", err)
	}
	defer tx.Rollback()

	q := sqlcdb.New(tx)

	newVersion, err := acquireVersion(ctx, q, merchantID, evt)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal payload: %w", err))
	}

	meta, err := json.Marshal(evt.Metadata)
	if err != nil {
		return kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal metadata: %w", err))
	}
	metaRaw := json.RawMessage(meta)

	updatedBy := ctxkey.GetUserName(ctx)
	var updatedByPtr *string
	if updatedBy != "" {
		updatedByPtr = &updatedBy
	}

	if _, err = q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		MerchantID:       merchantID,
		AggregateType:    evt.AggregateType,
		AggregateID:      evt.AggregateUuid,
		AggregateVersion: newVersion,
		EventType:        evt.EventType,
		EventUuid:        evt.Uuid,
		Payload:          payload,
		Metadata:         &metaRaw,
		UpdatedBy:        updatedByPtr,
	}); err != nil {
		return kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
	}

	return tx.Commit()
}

// acquireVersion checks and increments the aggregate version.
// Version=0 means new aggregate; version>0 expects exact match in DB.
func acquireVersion(ctx context.Context, q *sqlcdb.Queries, merchantID int64, evt event.Event) (int64, error) {
	if evt.Version == 0 {
		if err := q.InsertAggregateVersion(ctx, sqlcdb.InsertAggregateVersionParams{
			AggregateType:  evt.AggregateType,
			AggregateID:    evt.AggregateUuid,
			MerchantID:     merchantID,
			CurrentVersion: 1,
		}); err != nil {
			return 0, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
		}
		return 1, nil
	}

	result, err := q.UpdateVersionIfMatch(ctx, sqlcdb.UpdateVersionIfMatchParams{
		AggregateType:  evt.AggregateType,
		AggregateID:    evt.AggregateUuid,
		MerchantID:     merchantID,
		CurrentVersion: int64(evt.Version),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, kerrors.NewInfrastructureError(kerrors.ErrVersionConflict, evt, fmt.Errorf("expected version %d", evt.Version))
		}
		return 0, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
	}
	return result.CurrentVersion, nil
}
