package mediator

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
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

func NewMediator() *Mediator {
	return &Mediator{handlers: make(map[event_types.EventType]Handler)}
}

func (m *Mediator) Register(eventType event_types.EventType, h Handler) {
	m.handlers[eventType] = h
}

func (m *Mediator) Dispatch(ctx context.Context, evt event.Event) (HandlerResult, error) {
	h, ok := m.handlers[evt.EventType]
	if !ok {
		return HandlerResult{}, errors.NewRuntimeError(errors.ErrHandlerNotFound, evt, nil, false, true)
	}
	return h.Handle(ctx, evt)
}
