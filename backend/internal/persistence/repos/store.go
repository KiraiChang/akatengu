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

// Store appends events to the event store with optimistic locking, inside the
// caller's transaction.
type Store interface {
	// Append writes evt as the next version of its aggregate, which must be at
	// evt.ExpectedVersion; a new aggregate is created when ExpectedVersion is 0.
	Append(ctx context.Context, evt event.Event) (int64, int64, error)
	// Create writes evt as version 1 of a new aggregate. Any existing version
	// record, including one at version 0, is a conflict (ADR-023 A2 rule 3).
	Create(ctx context.Context, evt event.Event) (int64, int64, error)
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
			// 找不到符合的版本紀錄時，只有新的 aggregate（expected_version 為 0）可以建立
			if evt.ExpectedVersion != 0 {
				return eventID, newVersion, kerrors.NewInfrastructureError(kerrors.ErrVersionConflict, evt, fmt.Errorf("version mismatch: expected %d", evt.ExpectedVersion))
			}
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

	eventID, err = s.insertEvent(ctx, evt, newVersion)
	return eventID, newVersion, err
}

func (s *sqlxStore) Create(ctx context.Context, evt event.Event) (int64, int64, error) {
	// Append cannot guarantee a new aggregate: its UPDATE also matches a
	// version record at 0. Check that no record exists at all, in the same
	// transaction, then insert.
	_, err := s.q.GetAggregateVersion(ctx, sqlcdb.GetAggregateVersionParams{
		AggregateType: evt.AggregateType,
		AggregateID:   evt.AggregateUuid,
		MerchantID:    evt.MerchantID,
	})
	if err == nil {
		return 0, 0, kerrors.NewInfrastructureError(kerrors.ErrVersionConflict, evt, fmt.Errorf("aggregate %s already exists", evt.AggregateUuid))
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, 0, kerrors.NewInfrastructureError(kerrors.ErrStoreReadFailed, evt, fmt.Errorf("get version: %w", err))
	}

	const newVersion int64 = 1
	if err := s.q.InsertAggregateVersion(ctx, sqlcdb.InsertAggregateVersionParams{
		AggregateType:  evt.AggregateType,
		AggregateID:    evt.AggregateUuid,
		MerchantID:     evt.MerchantID,
		CurrentVersion: newVersion,
	}); err != nil {
		return 0, newVersion, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
	}

	eventID, err := s.insertEvent(ctx, evt, newVersion)
	return eventID, newVersion, err
}

// insertEvent writes evt at version after the version record is updated.
func (s *sqlxStore) insertEvent(ctx context.Context, evt event.Event, version int64) (int64, error) {
	payload, err := json.Marshal(evt.Payload)
	if err != nil {
		return 0, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal payload: %w", err))
	}

	meta, err := json.Marshal(evt.Metadata)
	if err != nil {
		return 0, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, fmt.Errorf("marshal metadata: %w", err))
	}
	metaRaw := json.RawMessage(meta)

	// 寫入事件
	eventID, err := s.q.InsertEvent(ctx, sqlcdb.InsertEventParams{
		MerchantID:       evt.MerchantID,
		AggregateType:    evt.AggregateType,
		AggregateID:      evt.AggregateUuid,
		AggregateVersion: version,
		EventType:        evt.EventType,
		EventUuid:        evt.Uuid,
		Payload:          payload,
		Metadata:         &metaRaw,
		UpdatedBy:        &evt.UpdatedBy,
	})
	if err != nil {
		return 0, kerrors.NewInfrastructureError(kerrors.ErrStoreWriteFailed, evt, err)
	}
	return eventID, nil
}
