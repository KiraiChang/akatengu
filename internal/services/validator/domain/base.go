package domain

import (
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/repos/query"
	"akatengu/internal/services/validator"
	"context"
	"encoding/json"
)

type domainValidator struct {
	query *query.Repo
	rules map[string]domainValidateFn
}

type domainValidateFn func(ctx context.Context, payload json.RawMessage) error

func NewValidator(query *query.Repo) validator.Validator {
	v := &domainValidator{
		query: query,
		rules: map[string]domainValidateFn{},
	}
	v.register()
	return v
}

func (d *domainValidator) Validate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) error {
	fn, ok := d.rules[eventType.String()]
	if !ok {
		return nil
	}
	return fn(ctx, payload)
}

func (d *domainValidator) register() {
	// ─────────────────────────────────────────
	// transaction
	// ─────────────────────────────────────────
	d.rules[event_types.EventTransactionCreated.String()] = d.validateEventTransactionCreated

	// ─────────────────────────────────────────
	// period close
	// ─────────────────────────────────────────

	// month closed.*
	d.rules[event_types.EventPeriodMonthClosed.String()] = d.validateEventPeriodMonthClosed

	// annul close.*
	d.rules[event_types.EventPeriodAnnualClosed.String()] = d.validateEventPeriodAnnualClosed
}
