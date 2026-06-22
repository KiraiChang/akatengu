package mediator

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"context"

	"github.com/jmoiron/sqlx"
)

type Handler interface {
	Handle(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error)
}

type HandlerFunc func(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error)

func (f HandlerFunc) Handle(ctx context.Context, repo *query.Repo, evt event.Event) (result.EventResult, error) {
	return f(ctx, repo, evt)
}

type Mediator struct {
	handlers map[event_types.EventType]Handler
	repo     *query.Repo
}

func (m *Mediator) WithQuery(db *sqlx.DB) *Mediator {
	m.repo = query.NewQueryRepository(db)
	return m
}

func NewMediator() *Mediator {
	return &Mediator{
		handlers: make(map[event_types.EventType]Handler),
	}
}

func (m *Mediator) Register(eventType event_types.EventType, h Handler) {
	m.handlers[eventType] = h
}

func (m *Mediator) Dispatch(ctx context.Context, evt event.Event) (result.EventResult, error) {
	h, ok := m.handlers[evt.EventType]
	if !ok {
		return result.EventResult{}, errors.NewRuntimeError(errors.ErrHandlerNotFound, evt)
	}
	return h.Handle(ctx, m.repo, evt)
}
