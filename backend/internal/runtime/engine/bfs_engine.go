package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/safety"
	"context"
)

type BFSEngine struct {
	process safety.BatchProcessor
}

func NewBFSEngine(executor *Executor, middlewares ...safety.Middleware) *BFSEngine {
	terminal := safety.BatchProcessor(func(ctx context.Context, _ int, batch []event.Event) ([]event.Event, error) {
		return executor.ExecuteBatch(ctx, batch)
	})
	return &BFSEngine{
		process: safety.Chain(terminal, middlewares...),
	}
}

func (e *BFSEngine) Run(ctx context.Context, events []event.Event) error {
	q := &Queue{}
	q.PushBatch(events)
	depth := 0

	for !q.Empty() {
		batch := q.PopCurrentLevel()
		next, err := e.process(ctx, depth, batch)
		if err != nil {
			return err
		}
		q.PushBatch(next)
		depth++
	}
	return nil
}
