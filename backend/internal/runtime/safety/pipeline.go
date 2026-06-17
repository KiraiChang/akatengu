package safety

import (
	"akatengu/internal/kernel/event"
	"context"
)

// BatchProcessor handles one BFS level and returns the next level's events.
type BatchProcessor func(ctx context.Context, depth int, batch []event.Event) ([]event.Event, error)

// Middleware wraps a BatchProcessor with pre/post logic.
type Middleware func(BatchProcessor) BatchProcessor

// Chain builds a pipeline from the provided middleware and a terminal processor.
// Middleware is applied in declaration order: the first entry is the outermost
// layer and therefore executes first at runtime.
func Chain(terminal BatchProcessor, middlewares ...Middleware) BatchProcessor {
	p := terminal
	for i := len(middlewares) - 1; i >= 0; i-- {
		p = middlewares[i](p)
	}
	return p
}
