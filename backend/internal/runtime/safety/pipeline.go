package safety

import (
	"akatengu/internal/kernel/event"
	"context"
)

// EventProcessor processes a single event and returns any child events it emits.
type EventProcessor func(ctx context.Context, depth int, evt event.Event) ([]event.Event, error)

// Middleware wraps an EventProcessor with pre/post logic.
type Middleware func(EventProcessor) EventProcessor

// Chain builds a pipeline from the provided middleware and a terminal processor.
// Middleware is applied in declaration order: the first entry is the outermost
// layer and therefore executes first at runtime.
func Chain(terminal EventProcessor, middlewares ...Middleware) EventProcessor {
	p := terminal
	for i := len(middlewares) - 1; i >= 0; i-- {
		p = middlewares[i](p)
	}
	return p
}
