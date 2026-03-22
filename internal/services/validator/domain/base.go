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
	rules map[string]validateFn
}

type validateFn func(ctx context.Context, payload json.RawMessage) error

func (v *domainValidator) Validate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) error {
	fn, ok := v.rules[eventType.String()]
	if !ok {
		return nil
	}
	return fn(ctx, payload)
}

func NewValidator(query *query.Repo) validator.Validator {
	domain := &domainValidator{
		query: query,
	}

	domain.register()

	return domain
}

func (v *domainValidator) register() {
	v.rules[event_types.EventPeriodClosed.String()] = v.validatePeriodClosed
}
