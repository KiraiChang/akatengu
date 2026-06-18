package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"context"
)

// DepthGuard returns a Middleware that aborts BFS expansion when depth
// exceeds maxDepth, preventing runaway handler chains.
// Error policy: Fatal=true (depth exceeded is unrecoverable).
func DepthGuard(maxDepth int) Middleware {
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) (result.EventResult, error) {
			if depth > maxDepth {
				return result.EventResult{Policy: result.Policy{Fatal: true}},
					kerrors.NewRuntimeError(kerrors.ErrBFSDepthExceeded, event.Event{})
			}
			return next(ctx, depth, evt)
		}
	}
}
