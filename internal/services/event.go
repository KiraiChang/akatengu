package services

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/calculator"
	"akatengu/internal/services/projection"
	"akatengu/internal/services/validator"
	"akatengu/internal/services/validator/domain"
	"akatengu/internal/services/validator/payload"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type AggregateState struct {
	Version int64
	Events  []db.EventStore // snapshot 之後的增量事件
	Snap    *db.Snapshot    // nil 代表沒有 snapshot
}

type EventStoreService struct {
	uow              event_store.UnitOfWork
	query            *query.Repo
	payloadValidator validator.Validator
	domainValidator  validator.Validator
	calculator       calculator.EventCalculator
	projections      []projection.Projection
}

func NewEventStoreService(uow event_store.UnitOfWork, query *query.Repo, projections []projection.Projection) *EventStoreService {
	return &EventStoreService{
		uow:              uow,
		query:            query,
		projections:      projections,
		payloadValidator: payload.NewValidator(),
		domainValidator:  domain.NewValidator(query),
		calculator:       calculator.NewCalculator(query),
	}
}

// Append 寫入一個事件，同步更新所有 projection
// 整個流程在同一個 transaction 內完成
func (es *EventStoreService) Append(ctx context.Context, cmd cmd.AppendCmd) (*db.EventStore, error) {
	payload, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	if err := es.payloadValidator.Validate(ctx, cmd.EventType, payload); err != nil {
		return nil, err // 直接回傳，不進 uow
	}

	if err := es.domainValidator.Validate(ctx, cmd.EventType, payload); err != nil {
		return nil, err // 直接回傳，不進 uow
	}

	payload, err = es.calculator.Calculate(ctx, cmd.EventType, payload)
	if err != nil {
		return nil, fmt.Errorf("calculate payload: %w", err)
	}

	var event db.EventStore
	err = es.uow.Do(ctx, func(tx event_store.EventStoreRepositories) error {

		// 版本控制
		newVersion, err := tx.Version.UpdateIfVersionMatch(ctx, cmd.AggregateType, cmd.AggregateID, cmd.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update version: %d fail, error: %w", cmd.ExpectedVersion, err)
		}

		// 寫入事件
		eventID, err := tx.Event.Insert(ctx, event_store.InsertEventParams{
			AggregateType:    cmd.AggregateType,
			AggregateID:      cmd.AggregateID,
			AggregateVersion: newVersion,
			EventType:        cmd.EventType,
			Payload:          payload,
		})
		if err != nil {
			return fmt.Errorf("insert event: %w", err)
		}

		event = db.EventStore{
			EventId:          eventID,
			AggregateType:    cmd.AggregateType,
			AggregateId:      cmd.AggregateID,
			AggregateVersion: newVersion,
			EventType:        cmd.EventType,
			Payload:          payload,
			OccurredAt:       time.Now(),
		}

		// 同步更新 projection
		for _, proj := range es.projections {
			if err := proj.Apply(ctx, tx, &event); err != nil {
				return fmt.Errorf("apply projection %s: %w", proj.Name(), err)
			}
		}

		// snapshot 決策
		if newVersion%50 == 0 {
			state, _ := json.Marshal(event)
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
	return &event, nil
}

// Replay 從指定 event_id 之後重播事件（用於重建 projection）
func (s *EventStoreService) Replay(ctx context.Context, fromEventID int64, aggregateType *enums.AggregateType) ([]db.EventStore, error) {
	events, err := s.query.Event.Replay(ctx, fromEventID, aggregateType)
	if err != nil {
		return nil, err
	}

	return events, nil
}
