package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/persistence/eventstore"
	"akatengu/internal/persistence/projection"
	"akatengu/internal/runtime/safety"
	"context"
)

type BFSEngine struct {
	executor    *Executor
	middlewares []safety.Middleware
	store       eventstore.Store
	projector   *projection.Registry
}

func NewBFSEngine(executor *Executor, middlewares ...safety.Middleware) *BFSEngine {
	return &BFSEngine{executor: executor, middlewares: middlewares}
}

func (e *BFSEngine) WithStore(s eventstore.Store) *BFSEngine {
	e.store = s
	return e
}

func (e *BFSEngine) WithProjector(r *projection.Registry) *BFSEngine {
	e.projector = r
	return e
}

func (e *BFSEngine) Run(ctx context.Context, evt event.Event) error {
	terminal := safety.EventProcessor(func(ctx context.Context, _ int, ev event.Event) ([]event.Event, error) {
		if e.store != nil {
			if err := e.store.Append(ctx, ev); err != nil {
				return nil, err
			}
		}

		result, err := e.executor.Dispatch(ctx, ev)
		if err != nil {
			return nil, err
		}

		if e.projector != nil {
			if err := e.projector.ApplyAll(ctx, ev, result.State); err != nil {
				return nil, err
			}
		}

		return result.Events, nil
	})

	process := safety.Chain(terminal, e.middlewares...)

	q := &Queue{}
	q.Push(evt, 0)

	for !q.Empty() {
		current, depth := q.Pop()
		children, err := process(ctx, depth, current)
		if err != nil {
			return err
		}
		for _, child := range children {
			q.Push(child, depth+1)
		}
	}
	return nil
}
