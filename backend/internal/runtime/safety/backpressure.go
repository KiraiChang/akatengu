package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
)

// BackpressureScheduler returns a Middleware that bounds total pending events
// in the BFS queue to maxPending.
//
//   - ErrBackpressureActive  – a root event (depth=0) enters when maxPending < 1.
//   - ErrQueueOverflow       – the children returned by a handler would push the
//     pending count past the limit.
func BackpressureScheduler(maxPending int) Middleware {
	pending := 0
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) ([]event.Event, error) {
			if depth == 0 {
				// New root event: reset pending counter to 1.
				pending = 1
				if pending > maxPending {
					return nil, kerrors.NewRuntimeError(kerrors.ErrBackpressureActive, evt, nil, true, false)
				}
			}

			children, err := next(ctx, depth, evt)
			if err != nil {
				return nil, err
			}

			// Current event done; account for children it emitted.
			pending = pending - 1 + len(children)
			if pending > maxPending {
				return nil, kerrors.NewRuntimeError(kerrors.ErrQueueOverflow, evt, nil, false, false)
			}

			return children, nil
		}
	}
}
