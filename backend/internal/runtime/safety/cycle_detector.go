package safety

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"context"
)

// CycleDetector returns a Middleware that detects duplicate event UUIDs within
// a single Run. The seen set is created at call time, so each call produces an
// independent guard instance.
// Error policy: Fatal=true (a UUID cycle is a programming error, not transient).
func CycleDetector() Middleware {
	seen := make(map[string]struct{})
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) (result.EventResult, error) {
			if _, ok := seen[evt.Uuid]; ok {
				return result.EventResult{Policy: result.Policy{Fatal: true}},
					kerrors.NewRuntimeError(kerrors.ErrEventLoopDetected, evt)
			}
			seen[evt.Uuid] = struct{}{}
			return next(ctx, depth, evt)
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
// Error policy: Fatal=true (a causal loop is a programming error).
func DeterministicLoopDetector() Middleware {
	graph := make(map[string]eventNode) // uuid → (eventType, causationID)
	return func(next EventProcessor) EventProcessor {
		return func(ctx context.Context, depth int, evt event.Event) (result.EventResult, error) {
			graph[evt.Uuid] = eventNode{
				eventType:   evt.EventType,
				causationID: evt.Metadata.Tracing.CausationID,
			}
			if err := walkCausationChain(graph, evt); err != nil {
				return result.EventResult{Policy: result.Policy{Fatal: true}}, err
			}
			return next(ctx, depth, evt)
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
			return kerrors.NewRuntimeError(kerrors.ErrEventLoopDetected, evt)
		}
		ancestorID = node.causationID
	}
	return nil
}
