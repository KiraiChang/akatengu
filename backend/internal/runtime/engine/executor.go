package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
	"context"
)

type Executor struct {
	mediator *mediator.Mediator
}

func NewExecutor(m *mediator.Mediator) *Executor {
	return &Executor{m}
}

func (e *Executor) ExecuteBatch(ctx context.Context, batch []event.Event) ([]event.Event, error) {
	var next []event.Event
	for _, evt := range batch {
		result, err := e.mediator.Dispatch(ctx, evt)
		if err != nil {
			return nil, err
		}
		next = append(next, result.Children...)
	}
	return next, nil
}
