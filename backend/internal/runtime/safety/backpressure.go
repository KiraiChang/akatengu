package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
)

// BackpressureScheduler returns a Middleware that bounds total pending events
// in the BFS queue to maxPending.
//
//   - ErrBackpressureActive  – the initial batch at depth 0 already exceeds the limit.
//   - ErrQueueOverflow       – the children returned by a handler would push the
//     pending count past the limit.
func BackpressureScheduler(maxPending int) Middleware {
	pending := 0
	return func(next BatchProcessor) BatchProcessor {
		return func(ctx context.Context, depth int, batch []event.Event) ([]event.Event, error) {
			if depth == 0 {
				pending = len(batch)
				if pending > maxPending {
					return nil, kerrors.NewRuntimeError(kerrors.ErrBackpressureActive, event.Event{}, nil, true, false)
				}
			}

			children, err := next(ctx, depth, batch)
			if err != nil {
				return nil, err
			}

			// Batch has been processed: remove it and account for children.
			pending = pending - len(batch) + len(children)
			if pending > maxPending {
				return nil, kerrors.NewRuntimeError(kerrors.ErrQueueOverflow, event.Event{}, nil, false, false)
			}

			return children, nil
		}
	}
}
