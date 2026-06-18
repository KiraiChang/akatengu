package repos

import (
	"akatengu/internal/database/sqlcdb"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// Store appends a single event to the event store with optimistic locking.
// Each Append call opens its own transaction.
type Store interface {
	Append(ctx context.Context, evt event.Event) (int64, int64, error)
}

type sqlxStore struct {
	q *sqlcdb.Queries
}

func NewStore(db *sqlx.Tx) Store {
	return &sqlxStore{q: sqlcdb.New(db)}
}

func (s *sqlxStore) Append(ctx context.Context, evt event.Event) (int64, int64, error) {
	var eventID, newVersion int64
	// 版本控制
	result, err := s.q.UpdateVersionIfMatch(ctx, sqlcdb.UpdateVersionIfMatchParams{
		AggregateType:  evt.AggregateType,
		AggregateID:    evt.AggregateUuid,
		MerchantID:     evt.MerchantID,
		CurrentVersion: evt.ExpectedVersion,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			newVersion = evt.ExpectedVersion + 1
			err := s.q.InsertAggregateVersion(ctx, sqlcdb.InsertAggregateVersionParams{
				AggregateType:  evt.AggregateType,
				AggregateID:    evt.AggregateUuid,
				MerchantID:     evt.MerchantID,
				CurrentVersion: newVersion,
			})
			if err != nil {
				return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
			}
		} else {
			return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("update version: %d fail, error: %w", evt.ExpectedVersion, err))
		}
	} else {
		newVersion = result.CurrentVersion
	}

	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal payload: %w", err))
	}

	meta, err := json.Marshal(evt.Metadata)
	if err != nil {
		return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal metadata: %w", err))
	}
	metaRaw := json.RawMessage(meta)

	// 寫入事件
	if _, err = s.q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		MerchantID:       evt.MerchantID,
		AggregateType:    evt.AggregateType,
		AggregateID:      evt.AggregateUuid,
		AggregateVersion: newVersion,
		EventType:        evt.EventType,
		EventUuid:        evt.Uuid,
		Payload:          payload,
		Metadata:         &metaRaw,
		UpdatedBy:        &evt.UpdatedBy,
	}); err != nil {
		return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
	}

	return eventID, newVersion, nil
}
