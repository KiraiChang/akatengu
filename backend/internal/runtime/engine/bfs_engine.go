package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/runtime/safety"
	"context"
)

type BFSEngine struct {
	executor  *Executor
	process   safety.EventProcessor
	uow       repos.UnitOfWork
	projector *handle.Registry
}

func NewBFSEngine(executor *Executor) *BFSEngine {
	e := &BFSEngine{executor: executor}
	e.process = func(ctx context.Context, _ int, ev event.Event) (result.EventResult, error) {
		r, err := e.executor.Dispatch(ctx, ev)
		if err != nil {
			// Pass through the result so callers can inspect r.Policy.
			return r, err
		}

		if e.uow != nil {
			err = e.uow.Do(ctx, func(tx *repos.Repos) error {
				_, _, err := tx.Store.Append(ctx, ev)
				if err != nil {
					return err
				}
				if e.projector != nil {
					if err := e.projector.ApplyAll(ctx, ev, r.State); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return r, err
			}
		}

		return r, nil
	}
	return e
}

func (e *BFSEngine) WithMiddleware(middlewares ...safety.Middleware) *BFSEngine {
	e.process = safety.Chain(e.process, middlewares...)
	return e
}

func (e *BFSEngine) WithUnitOfWork(uow repos.UnitOfWork) *BFSEngine {
	e.uow = uow
	return e
}

func (e *BFSEngine) WithProjector(r *handle.Registry) *BFSEngine {
	e.projector = r
	return e
}

func (e *BFSEngine) Run(ctx context.Context, evt event.Event) error {
	q := &Queue{}
	q.Push(evt, 0)

	for !q.Empty() {
		current, depth := q.Pop()
		r, err := e.process(ctx, depth, current)
		if err != nil {
			// r.Policy describes the error-handling strategy:
			//   Fatal=true     → unrecoverable, abort immediately.
			//   Retryable=true → transient, eligible for retry (retry logic TBD).
			// Currently all errors abort the run regardless of policy.
			return err
		}
		for _, child := range r.Events {
			q.Push(child, depth+1)
		}
	}
	return nil
}
