package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
)

// CycleDetector returns a Middleware that detects duplicate event UUIDs.
// The seen set is created when CycleDetector() is called, so each call
// produces an independent guard instance with its own empty set.
func CycleDetector() Middleware {
	seen := make(map[string]struct{})
	return func(next BatchProcessor) BatchProcessor {
		return func(ctx context.Context, depth int, batch []event.Event) ([]event.Event, error) {
			for _, evt := range batch {
				if _, ok := seen[evt.Uuid]; ok {
					return nil, kerrors.NewRuntimeError(kerrors.ErrEventLoopDetected, evt, nil, false, true)
				}
				seen[evt.Uuid] = struct{}{}
			}
			return next(ctx, depth, batch)
		}
	}
}
