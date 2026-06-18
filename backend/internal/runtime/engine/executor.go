package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
	"context"
	"sync"
)

type Executor struct {
	mediator *mediator.Mediator
}

func NewExecutor(m *mediator.Mediator) *Executor {
	return &Executor{m}
}

// ExecuteBatch dispatches all events in the batch concurrently and returns their
// children in parent-insertion order.
//
// Results are written into per-index slots (slots[i] = handler i's children) and
// flattened in index order after all goroutines finish. This separates the "async
// dispatch" concern from the "ordering" concern: the flatten step never changes
// regardless of goroutine completion order.
//
// Error handling: all goroutines always run to completion (wg.Wait). The first
// error encountered by index order is returned; remaining errors are discarded.
// Context cancellation is forwarded to each Dispatch call for cooperative abort.
func (e *Executor) ExecuteBatch(ctx context.Context, batch []event.Event) ([]event.Event, error) {
	type slot struct {
		children []event.Event
		err      error
	}
	slots := make([]slot, len(batch))

	var wg sync.WaitGroup
	for i, evt := range batch {
		i, evt := i, evt
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := e.mediator.Dispatch(ctx, evt)
			slots[i] = slot{r.Children, err}
		}()
	}
	wg.Wait()

	var next []event.Event
	for _, s := range slots {
		if s.err != nil {
			return nil, s.err
		}
		next = append(next, s.children...)
	}
	return next, nil
}
