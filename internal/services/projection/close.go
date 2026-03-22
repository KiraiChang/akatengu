package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"encoding/json"
	"time"
)

type ClosingProjectionService struct{}

func (s *ClosingProjectionService) Name() string { return enums.AggreagateInvestment.String() }

func (s *ClosingProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, prevResult Result) (Result, error) {
	switch event.EventType.String() {
	case event_types.EventPeriodCloseStarted.String():
		return s.applyCloseStarted(ctx, tx, event, prevResult)
	case event_types.EventPeriodClosed.String():
		return s.applyClosed(ctx, tx, event, prevResult)
	case event_types.EventPeriodReopened.String():
		return s.applyReopened(ctx, tx, event, prevResult)
	}
	return prevResult, nil
}

func (p *ClosingProjectionService) applyCloseStarted(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var payload payload.PeriodCloseStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return result, err
	}

	// 建立 period_closings 紀錄，status = open
	if err := tx.Projection.InsertClose(ctx, db.PeriodClosing{
		ClosingId:   event.EventId,
		PeriodType:  payload.PeriodType,
		PeriodStart: payload.PeriodStart,
		PeriodEnd:   payload.PeriodEnd,
		Status:      enums.ClosingStatusOpen,
	}); err != nil {
		return result, err
	}

	return result, nil
}

func (p *ClosingProjectionService) applyClosed(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var payload payload.PeriodClosedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return result, err
	}

	//if err := tx.Projection.UpdateCloseSnapshotByPeriod(ctx,
	//	payload.PeriodType,
	//	payload.PeriodStart,
	//); err != nil {
	//	return result, err
	//}

	return result, nil
}

func (p *ClosingProjectionService) applyReopened(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore, result Result) (Result, error) {
	var payload payload.PeriodReopenedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return result, err
	}

	return result, nil
}

func marshalSnapshot(data db.SnapshotData) (string, error) {
	data.GeneratedAt = time.Now().Format(time.RFC3339)
	b, err := json.Marshal(data)
	return string(b), err
}
