package handle

import (
	event_payload "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	dbprojection "akatengu/internal/model/db/projection"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/shared/utils"
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// NewPeriodProjector opens periods in period_closings, the same rows the legacy
// PeriodProjectionService (services/projection/period.go) writes. It only
// handles the opening events; closing and reopening stay on the legacy path
// until R5 (BFS R2 D10).
func NewPeriodProjector(db *sqlx.DB) Projector {
	c := &periodProjector{
		handles: make(map[event_types.EventType]handle),
	}
	c.handles[event_types.EventPeriodMonthStarted.Enum()] = c.applyMonthStarted
	c.handles[event_types.EventPeriodAnnualStarted.Enum()] = c.applyAnnualStarted
	return c
}

type periodProjector struct {
	handles map[event_types.EventType]handle
}

func (c periodProjector) EventTypes() []event_types.EventType {
	return utils.Keys(c.handles)
}

func (c periodProjector) Apply(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	h, ok := c.handles[evt.EventType]
	if !ok {
		return errors.NewRuntimeError(errors.ErrProjectorNotFound, evt)
	}
	return h(ctx, tx, evt)
}

func (c periodProjector) Name() string {
	return "periodProjector"
}

// applyMonthStarted opens the month as an OPEN period named after the event
// UUID, as the legacy projection derives closing_uuid (BFS R2 D12).
func (c *periodProjector) applyMonthStarted(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.PeriodMonthStartedPayload](&evt)
	if err != nil {
		return err
	}
	t, err := time.Parse("2006-01", p.PeriodStart)
	if err != nil {
		return fmt.Errorf("invalid year-month %q: %w", p.PeriodStart, err)
	}
	first := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return c.open(ctx, tx, evt, enums.PeriodMonthly.Enum(), first, first.AddDate(0, 1, -1))
}

func (c *periodProjector) applyAnnualStarted(ctx context.Context, tx *repos.DbTransaction, evt event.Event) error {
	p, err := checkAndGetPayload[event_payload.PeriodAnnualStartedPayload](&evt)
	if err != nil {
		return err
	}
	first := time.Date(p.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	return c.open(ctx, tx, evt, enums.PeriodAnnual.Enum(), first, first.AddDate(1, 0, -1))
}

// open inserts the period from first to last; opening the same period twice
// fails on the (merchant_id, period_type, period_start) unique key.
func (c *periodProjector) open(ctx context.Context, tx *repos.DbTransaction, evt event.Event,
	periodType enums.PeriodType, first, last time.Time) error {
	return tx.Projection.InsertPeriodClose(ctx, dbprojection.PeriodClosing{
		MerchantID:  evt.MerchantID,
		ClosingUuid: evt.Uuid,
		PeriodType:  periodType,
		PeriodStart: first.Format("2006-01-02"),
		PeriodEnd:   last.Format("2006-01-02"),
		Status:      enums.PeriodTypeStatusOpen.Enum(),
		UpdatedBy:   toUpdatedBy(evt.UpdatedBy),
	})
}
