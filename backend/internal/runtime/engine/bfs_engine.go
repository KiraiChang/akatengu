package engine

import (
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
)

type BFSEngine struct {
	executor *Executor
	maxDepth int
}

func NewBFSEngine(executor *Executor, maxDepth int) *BFSEngine {
	return &BFSEngine{executor: executor, maxDepth: maxDepth}
}

func (e *BFSEngine) Run(ctx context.Context, events []event.Event) error {
	q := &Queue{}
	q.PushBatch(events)
	seen := make(map[string]struct{})
	depth := 0

	for !q.Empty() {
		if depth > e.maxDepth {
			return errors.NewRuntimeError(errors.ErrBFSDepthExceeded, event.Event{}, nil, false, true)
		}
		batch := q.PopCurrentLevel()
		for _, evt := range batch {
			if _, ok := seen[evt.Uuid]; ok {
				return errors.NewRuntimeError(errors.ErrEventLoopDetected, evt, nil, false, true)
			}
			seen[evt.Uuid] = struct{}{}
		}
		next, err := e.executor.ExecuteBatch(ctx, batch)
		if err != nil {
			return err
		}
		q.PushBatch(next)
		depth++
	}
	return nil
}
