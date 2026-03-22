package calculator

import (
	"akatengu/internal/model/enums/event_types"
	"encoding/json"
)

type EventCalculator interface {
	Calculate(eventType event_types.EventType, payload json.RawMessage) (json.RawMessage, error)
}

type eventCalculator struct {
	rules map[string]calculatorFn
}

type calculatorFn func(payload json.RawMessage) (json.RawMessage, error)

func NewCalculator() EventCalculator {
	v := &eventCalculator{
		rules: map[string]calculatorFn{},
	}
	v.register()
	return v
}

func (v *eventCalculator) Calculate(eventType event_types.EventType, payload json.RawMessage) (json.RawMessage, error) {
	fn, ok := v.rules[eventType.String()]
	if !ok {
		// 沒有 key 代表不用做計算
		return payload, nil
	}
	return fn(payload)
}

func (v *eventCalculator) register() {
	//v.rules[event_types.EventPeriodClosed] =
}
