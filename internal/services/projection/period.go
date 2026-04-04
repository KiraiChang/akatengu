package projection

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/payload/state"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services/pipelines"
	"context"
	"fmt"
	"time"
)

type PeriodProjectionService struct{}

func (s *PeriodProjectionService) Name() string { return "PERIOD_CLOSED" }

func (s *PeriodProjectionService) Apply(ctx context.Context, tx event_store.EventStoreRepositories, t event_types.EventType, ct *pipelines.Result) error {
	switch t.String() {
	case event_types.EventPeriodMonthStarted.String():
		return s.applyMonthStarted(ctx, tx, ct)
	case event_types.EventPeriodMonthClosed.String():
		return s.applyMonthClosed(ctx, tx, ct)
	case event_types.EventPeriodMonthReopened.String():
		return s.applyMonthReopened(ctx, tx, ct)
	case event_types.EventPeriodAnnualStarted.String():
		return s.applyAnnualStarted(ctx, tx, ct)
	case event_types.EventPeriodAnnualClosed.String():
		return s.applyAnnualClosed(ctx, tx, ct)
	case event_types.EventPeriodAnnualReopened.String():
		return s.applyAnnualReopened(ctx, tx, ct)
	}
	return nil
}

func (s *PeriodProjectionService) applyMonthStarted(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodMonthStartedPayload](ct)
	if err != nil {
		return err
	}

	first, last, err := s.periodMonthRange(p.PeriodStart)
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

func (s *PeriodProjectionService) applyMonthClosed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodMonthClosedPayload](ct)
	if err != nil {
		return err
	}

	state, err := checkAndGetState[state.PeriodMonthClosedState](ct)
	if err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.UpdatePeriodCloseSnapshotByPeriod(ctx,
		p.ClosingId,
		state.Snapshot,
		&p.ClosedAt,
	); err != nil {
		return err
	}

	// 建立 下一期 period_closings 紀錄，status = open
	if state.Next != nil {
		_, err := tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, *state.Next)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PeriodProjectionService) applyMonthReopened(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodMonthReopenedPayload](ct)
	if err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.ReopenPeriodClose(ctx,
		p.ClosingId,
		p.Reason,
		p.ReopenedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PeriodProjectionService) applyAnnualStarted(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodAnnualStartedPayload](ct)
	if err != nil {
		return err
	}

	first, last, err := s.periodAnnualRange(p.Year)
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

func (s *PeriodProjectionService) applyAnnualClosed(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodAnnualClosedPayload](ct)
	if err != nil {
		return err
	}

	state, err := checkAndGetState[state.PeriodAnnualClosedState](ct)
	if err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.UpdatePeriodCloseSnapshotByPeriod(ctx,
		p.ClosingId,
		state.Snapshot,
		&p.ClosedAt,
	); err != nil {
		return err
	}

	// 建立 下一期 period_closings 紀錄，status = open
	if state.Next != nil {
		_, err := tx.Projection.PeriodCloseRepo.InsertPeriodClose(ctx, *state.Next)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PeriodProjectionService) applyAnnualReopened(ctx context.Context, tx event_store.EventStoreRepositories, ct *pipelines.Result) error {
	p, err := checkAndGetPayload[payload.PeriodAnnualReopenedPayload](ct)
	if err != nil {
		return err
	}

	if err := tx.Projection.PeriodCloseRepo.ReopenPeriodClose(ctx,
		p.ClosingId,
		p.Reason,
		p.ReopenedAt,
	); err != nil {
		return err
	}

	return nil
}

func (s *PeriodProjectionService) nextPeriodMonthRange(date string) (startDate, endDate string, err error) {
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

func (s *PeriodProjectionService) periodMonthRange(yearMonth string) (startDate, endDate string, err error) {
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

func (s *PeriodProjectionService) periodAnnualRange(year int) (startDate, endDate string, err error) {
	// 當月第一天
	first := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// 當月最後一天（下個月第一天 -1 天）
	last := first.AddDate(1, 0, -1)

	return first.Format("2006-01-02"), last.Format("2006-01-02"), nil
}

func (s *PeriodProjectionService) nextPeriodAnnualRange(startDay string) (startDate, endDate string, err error) {
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
