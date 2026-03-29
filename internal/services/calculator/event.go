package calculator

import (
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/repos/query"
	"context"
	"encoding/json"
)

type EventCalculator interface {
	Calculate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) (json.RawMessage, error)
}

type eventCalculator struct {
	query *query.Repo
	rules map[string]calculatorFn
}

type calculatorFn func(ctx context.Context, payload json.RawMessage) (json.RawMessage, error)

func NewCalculator(query *query.Repo) EventCalculator {
	c := &eventCalculator{
		query: query,
		rules: map[string]calculatorFn{},
	}
	c.register()
	return c
}

func (e *eventCalculator) Calculate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) (json.RawMessage, error) {
	fn, ok := e.rules[eventType.String()]
	if !ok {
		// 沒有 key 代表不用做計算
		return payload, nil
	}
	return fn(ctx, payload)
}

func (e *eventCalculator) register() {

	// ─────────────────────────────────────────
	// period close
	// ─────────────────────────────────────────

	// month close.*
	e.rules[event_types.EventPeriodMonthClosed.String()] = e.calEventPeriodMonthClosed

	// annul close.*
	e.rules[event_types.EventPeriodAnnualClosed.String()] = e.calEventPeriodAnnualClosed
	e.rules[event_types.EventPeriodAnnualReopened.String()] = e.calEventPeriodAnnualReopened
}
