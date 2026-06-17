package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"context"
)

// CycleDetector returns a Middleware that detects duplicate event UUIDs within
// a single Run. The seen set is created at call time, so each call produces an
// independent guard instance.
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

// eventNode is an internal record used by DeterministicLoopDetector.
type eventNode struct {
	eventType   event_types.EventType
	causationID string
}

// DeterministicLoopDetector returns a Middleware that detects EventType cycles
// by traversing the causation chain embedded in each event's metadata.
// Unlike CycleDetector (UUID dedup) or DepthGuard (depth counter), this guard
// uses the actual causal graph: if processing EventType A produces a descendant
// of EventType A anywhere in its lineage, the cycle is detected and aborted.
func DeterministicLoopDetector() Middleware {
	graph := make(map[string]eventNode) // uuid → (eventType, causationID)
	return func(next BatchProcessor) BatchProcessor {
		return func(ctx context.Context, depth int, batch []event.Event) ([]event.Event, error) {
			for _, evt := range batch {
				graph[evt.Uuid] = eventNode{
					eventType:   evt.EventType,
					causationID: evt.Metadata.Tracing.CausationID,
				}
				if err := walkCausationChain(graph, evt); err != nil {
					return nil, err
				}
			}
			return next(ctx, depth, batch)
		}
	}
}

// walkCausationChain walks up the CausationID chain to find whether evt's
// EventType already appears among its ancestors, indicating a deterministic cycle.
func walkCausationChain(graph map[string]eventNode, evt event.Event) error {
	visited := make(map[string]struct{})
	ancestorID := evt.Metadata.Tracing.CausationID
	for ancestorID != "" {
		if _, seen := visited[ancestorID]; seen {
			break // guard against malformed chains
		}
		visited[ancestorID] = struct{}{}
		node, ok := graph[ancestorID]
		if !ok {
			break
		}
		if node.eventType == evt.EventType {
			return kerrors.NewRuntimeError(kerrors.ErrEventLoopDetected, evt, nil, false, true)
		}
		ancestorID = node.causationID
	}
	return nil
}
