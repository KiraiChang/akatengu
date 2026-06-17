package engine

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"context"
	"fmt"
)

type HandlerResult struct {
	Children []event.Event
}

type Handler interface {
	Handle(ctx context.Context, evt event.Event) (HandlerResult, error)
}

type HandlerFunc func(ctx context.Context, evt event.Event) (HandlerResult, error)

func (f HandlerFunc) Handle(ctx context.Context, evt event.Event) (HandlerResult, error) {
	return f(ctx, evt)
}

type Mediator struct {
	handlers map[event_types.EventType]Handler
}

func newMediator() *Mediator {
	return &Mediator{handlers: make(map[event_types.EventType]Handler)}
}

func (m *Mediator) register(eventType event_types.EventType, h Handler) {
	m.handlers[eventType] = h
}

func (m *Mediator) dispatch(ctx context.Context, evt event.Event) (HandlerResult, error) {
	h, ok := m.handlers[evt.EventType]
	if !ok {
		return HandlerResult{}, fmt.Errorf("eventType=%s: %w", evt.EventType, ErrHandlerNotFound)
	}
	return h.Handle(ctx, evt)
}

type Executor struct {
	mediator *Mediator
}

func NewExecutor() *Executor {
	return &Executor{mediator: newMediator()}
}

func (e *Executor) Register(eventType event_types.EventType, h Handler) *Executor {
	e.mediator.register(eventType, h)
	return e
}

func (e *Executor) ExecuteBatch(ctx context.Context, batch []event.Event) ([]event.Event, error) {
	var next []event.Event
	for _, evt := range batch {
		result, err := e.mediator.dispatch(ctx, evt)
		if err != nil {
			return nil, err
		}
		next = append(next, result.Children...)
	}
	return next, nil
}
