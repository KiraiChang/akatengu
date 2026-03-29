package projection

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/unit_of_work/event_store"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type PeriodProjectionService struct{}

func (p *PeriodProjectionService) Name() string { return enums.AggregateTransaction.String() }

func (p *PeriodProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	switch event.EventType.String() {
	case event_types.EventPeriodMonthStarted.String():
		return p.applyMonthStarted(ctx, tx, event)
	case event_types.EventPeriodMonthClosed.String():
		return p.applyMonthClosed(ctx, tx, event)
	case event_types.EventPeriodMonthReopened.String():
		return p.applyMonthReopened(ctx, tx, event)
	case event_types.EventPeriodAnnualStarted.String():
		return p.applyAnnualStarted(ctx, tx, event)
	case event_types.EventPeriodAnnualClosed.String():
		return p.applyAnnualClosed(ctx, tx, event)
	case event_types.EventPeriodAnnualReopened.String():
		return p.applyAnnualReopened(ctx, tx, event)
	}
	return nil
}

func (p *PeriodProjectionService) applyMonthStarted(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodMonthStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	first, last, err := p.periodMonthRange(payload.PeriodStart)
	if err != nil {
		return err
	}

	// 建立 period_closings 紀錄，status = open
	_, err = tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, projection.PeriodClosing{
		PeriodType:  enums.PeriodMonthly,
		PeriodStart: first,
		PeriodEnd:   last,
		Status:      enums.PeriodTypeStatusOpen,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *PeriodProjectionService) applyMonthClosed(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodMonthClosedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.UpdatePeriodCloseSnapshotByPeriod(ctx,
		payload.ClosingId,
		payload.Snapshot,
		&payload.ClosedAt,
	); err != nil {
		return err
	}

	// 建立 下一期 period_closings 紀錄，status = open
	if payload.Next != nil {
		_, err := tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, *payload.Next)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PeriodProjectionService) applyMonthReopened(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodMonthReopenedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.ReopenPeriodClose(ctx,
		payload.ClosingId,
		payload.Reason,
		payload.ReopenedAt,
	); err != nil {
		return err
	}

	return nil
}

func (p *PeriodProjectionService) applyAnnualStarted(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodAnnualStartedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	first, last, err := p.periodAnnualRange(payload.Year)
	if err != nil {
		return err
	}

	// 建立 period_closings 紀錄，status = open
	_, err = tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, projection.PeriodClosing{
		PeriodType:  enums.PeriodAnnual,
		PeriodStart: first,
		PeriodEnd:   last,
		Status:      enums.PeriodTypeStatusOpen,
	})
	if err != nil {
		return err
	}

	return nil
}

func (p *PeriodProjectionService) applyAnnualClosed(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodAnnualClosedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.UpdatePeriodCloseSnapshotByPeriod(ctx,
		payload.ClosingId,
		payload.Snapshot,
		&payload.ClosedAt,
	); err != nil {
		return err
	}

	// 建立 下一期 period_closings 紀錄，status = open
	if payload.Next != nil {
		_, err := tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, *payload.Next)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PeriodProjectionService) applyAnnualReopened(ctx context.Context, tx event_store.EventStoreRepositories, event *db.EventStore) error {
	var payload payload.PeriodAnnualReopenedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.ReopenPeriodClose(ctx,
		payload.ClosingId,
		payload.Reason,
		payload.ReopenedAt,
	); err != nil {
		return err
	}

	return nil
}

func (p *PeriodProjectionService) nextPeriodMonthRange(date string) (startDate, endDate string, err error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "", "", fmt.Errorf("invalid date %q: %w", date, err)
	}

	// 下個月第一天
	first := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)

	// 下個月最後一天
	last := first.AddDate(0, 1, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}

func (p *PeriodProjectionService) periodMonthRange(yearMonth string) (startDate, endDate string, err error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return "", "", fmt.Errorf("invalid year-month %q: %w", yearMonth, err)
	}

	// 當月第一天
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(0, 1, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}

func (p *PeriodProjectionService) periodAnnualRange(year int) (startDate, endDate string, err error) {
	// 當月第一天
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(1, 0, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}

func (p *PeriodProjectionService) nextPeriodAnnualRange(startDay string) (startDate, endDate string, err error) {
	t, err := time.Parse("2006-01-02", startDay)
	if err != nil {
		return "", "", fmt.Errorf("invalid year-month %q: %w", startDay, err)
	}
	// 當月第一天
	first := time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(1, 0, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}
