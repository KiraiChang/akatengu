package services

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"akatengu/internal/services/pipelines/factory"
	"akatengu/internal/services/projection"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type AggregateState struct {
	Version int64
	Events  []db.EventStore // snapshot 之後的增量事件
	Snap    *db.Snapshot    // nil 代表沒有 snapshot
}

type EventStoreService struct {
	uow         event_store.UnitOfWork
	query       *query.Repo
	factory     pipelines.PipelineRegistry
	projections []projection.Projection
}

func NewEventStoreService(uow event_store.UnitOfWork, query *query.Repo) *EventStoreService {
	return &EventStoreService{
		uow:         uow,
		query:       query,
		projections: projection.NewProjection(),
		factory:     factory.NewPipelineRegistry(query),
	}
}

// Append 寫入一個事件，同步更新所有 projection
// 整個流程在同一個 transaction 內完成
func (es *EventStoreService) Append(ctx context.Context, cmd cmd.AppendCmd) (*db.EventStore, error) {
	ct, err := es.factory.Dispatch(ctx, cmd)
	if err != nil {
		return nil, err
	}
	merchantID, err := ctxkey.GetMerchantID(ctx)
	if err != nil {
		return nil, err
	}
	ct.MerchantID = merchantID
	ct.UpdatedBy = ctxkey.GetUserName(ctx)

	err = es.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {

		// 版本控制
		newVersion, err := tx.Version.UpdateIfVersionMatch(ctx, cmd.AggregateType, cmd.AggregateID, cmd.ExpectedVersion)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				newVersion = cmd.ExpectedVersion + 1
				err = tx.Version.Insert(ctx, cmd.AggregateType, cmd.AggregateID, newVersion)
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("update version: %d fail, error: %w", cmd.ExpectedVersion, err)
			}
		}

		// 寫入事件
		var updatedBy *string
		if ct.UpdatedBy != "" {
			updatedBy = &ct.UpdatedBy
		}
		eventID, err := tx.Event.Insert(ctx, event_store.InsertEventParams{
			AggregateType:    cmd.AggregateType,
			AggregateID:      cmd.AggregateID,
			AggregateVersion: newVersion,
			EventType:        cmd.EventType,
			Payload:          cmd.Payload,
			UpdatedBy:        updatedBy,
		})
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		ct.Event = db.EventStore{
			EventId:          eventID,
			AggregateType:    cmd.AggregateType,
			AggregateId:      cmd.AggregateID,
			AggregateVersion: newVersion,
			EventType:        cmd.EventType,
			Payload:          cmd.Payload,
			OccurredAt:       time.Now(),
		}

		// 同步更新 projection
		for _, proj := range es.projections {
			if err := proj.Apply(ctx, tx, cmd.EventType, ct); err != nil {
				return fmt.Errorf("apply projection %s: %w", proj.Name(), err)
			}
			if err = tx.Check.Upsert(ctx, proj.Name(), eventID); err != nil {
				return fmt.Errorf("upsert projection %s: %w", proj.Name(), err)
			}
		}

		// snapshot 決策
		if newVersion%50 == 0 {
			state, _ := json.Marshal(ct.Event)
			if err := tx.Snap.Upsert(ctx, db.Snapshot{
				AggregateType: cmd.AggregateType,
				AggregateId:   cmd.AggregateID,
				AtVersion:     newVersion,
				State:         state,
			}); err != nil {
				return fmt.Errorf("upsert snapshot: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return &ct.Event, nil
}

// Replay 重播事件以重建 projection。
// fromEventID == 0 代表全量重建：先清除所有 projection 資料再重播全部事件。
// fromEventID > 0 代表增量補播：不清除，直接從該 event_id 之後的事件繼續套用。
func (s *EventStoreService) Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error) {
	// 1. 取得待重播的事件
	events, err := s.query.Event.Replay(ctx, fromEventID, aggregateType)
	if err != nil {
		return nil, err
	}

	// 2. 全量重播：先清除 projection 資料並重置 checkpoint
	if fromEventID == 0 {
		if err := s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
			if err := tx.Truncate.TruncateProjections(ctx); err != nil {
				return err
			}
			seen := make(map[string]struct{})
			for _, proj := range s.projections {
				if _, ok := seen[proj.Name()]; ok {
					continue
				}
				seen[proj.Name()] = struct{}{}
				if err := tx.Check.Update(ctx, proj.Name(), 0); err != nil {
					return fmt.Errorf("reset checkpoint %s: %w", proj.Name(), err)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("truncate projections: %w", err)
		}
	}

	// 3. 逐一重播每個事件：執行 pipeline 取得 Result，再在 transaction 內套用所有 projection
	for _, event := range events {
		appCmd := cmd.AppendCmd{
			AggregateType:   event.AggregateType,
			AggregateID:     event.AggregateId,
			ExpectedVersion: event.AggregateVersion - 1,
			EventType:       event.EventType,
			Payload:         event.Payload,
		}

		ct, err := s.factory.Dispatch(ctx, appCmd)
		if err != nil {
			return nil, fmt.Errorf("dispatch event %d: %w", event.EventId, err)
		}
		ct.Event = event
		ct.MerchantID = event.MerchantID
		ct.UpdatedBy = *event.UpdatedBy
		if err := s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
			for _, proj := range s.projections {
				if err := proj.Apply(ctx, tx, event.EventType, ct); err != nil {
					return fmt.Errorf("projection %s: %w", proj.Name(), err)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("replay event %d: %w", event.EventId, err)
		}
	}

	// 4. 更新所有 projection 的 checkpoint 至最後一個重播的 event_id
	if len(events) > 0 {
		lastID := events[len(events)-1].EventId
		seen := make(map[string]struct{})
		if err := s.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {
			for _, proj := range s.projections {
				if _, ok := seen[proj.Name()]; ok {
					continue
				}
				seen[proj.Name()] = struct{}{}
				if err := tx.Check.Update(ctx, proj.Name(), lastID); err != nil {
					return fmt.Errorf("checkpoint %s: %w", proj.Name(), err)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("update checkpoints: %w", err)
		}
	}

	return events, nil
}
