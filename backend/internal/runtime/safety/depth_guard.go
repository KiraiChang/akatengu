package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"context"
)

// DepthGuard returns a Middleware that aborts BFS expansion when depth
// exceeds maxDepth, preventing runaway handler chains.
func DepthGuard(maxDepth int) Middleware {
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) ([]event.Event, error) {
			if depth > maxDepth {
				return nil, kerrors.NewRuntimeError(kerrors.ErrBFSDepthExceeded, event.Event{}, nil, false, true)
			}
			return next(ctx, depth, evt)
		}
	}
}
