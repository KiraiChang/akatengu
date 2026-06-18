package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"context"
)

// BackpressureScheduler returns a Middleware that bounds total pending events
// in the BFS queue to maxPending.
//
//   - ErrBackpressureActive  – a root event (depth=0) enters when maxPending < 1.
//     Policy: Retryable=true (caller may retry when load drops).
//   - ErrQueueOverflow       – the children returned by a handler would push the
//     pending count past the limit.
//     Policy: zero (not retryable, not fatal — caller decides).
func BackpressureScheduler(maxPending int) Middleware {
	pending := 0
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) (result.EventResult, error) {
			if depth == 0 {
				// New root event: reset pending counter to 1.
				pending = 1
				if pending > maxPending {
					return result.EventResult{Policy: result.Policy{Retryable: true}},
						kerrors.NewRuntimeError(kerrors.ErrBackpressureActive, evt)
				}
			}

			r, err := next(ctx, depth, evt)
			if err != nil {
				return r, err
			}

			// Current event done; account for children it emitted.
			pending = pending - 1 + len(r.Events)
			if pending > maxPending {
				return result.EventResult{}, kerrors.NewRuntimeError(kerrors.ErrQueueOverflow, evt)
			}

			return r, nil
		}
	}
}
