package engine

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
)

type testPayload struct{ typ event_types.EventType }

func (p testPayload) EventType() event_types.EventType { return p.typ }

var (
	typeA event_types.EventType
	typeB event_types.EventType
)

func TestMain(m *testing.M) {
	enums.InitEnums()
	typeA = event_types.EventTransactionCreated.Enum()
	typeB = event_types.EventTransactionCorrected.Enum()
	os.Exit(m.Run())
}

func newTestEvent(typ event_types.EventType) event.Event {
	return event.NewEvent[testPayload]("agg-1", testPayload{typ: typ})
}

func TestBFSEngine_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("SingleEvent_NoChildren", func(t *testing.T) {
		called := false
		exec := NewExecutor().Register(typeA, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
			called = true
			return HandlerResult{}, nil
		}))
		err := NewBFSEngine(exec, 10).Run(ctx, []event.Event{newTestEvent(typeA)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("handler was not called")
		}
	})

	t.Run("BFSExpansion_TwoLevels", func(t *testing.T) {
		var order []string
		exec := NewExecutor().
			Register(typeA, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
				order = append(order, "A")
				return HandlerResult{Children: []event.Event{newTestEvent(typeB)}}, nil
			})).
			Register(typeB, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
				order = append(order, "B")
				return HandlerResult{}, nil
			}))

		err := NewBFSEngine(exec, 10).Run(ctx, []event.Event{newTestEvent(typeA)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != 2 || order[0] != "A" || order[1] != "B" {
			t.Fatalf("expected [A B], got %v", order)
		}
	})

	t.Run("DepthExceeded", func(t *testing.T) {
		exec := NewExecutor().Register(typeA, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
			return HandlerResult{Children: []event.Event{newTestEvent(typeA)}}, nil
		}))
		err := NewBFSEngine(exec, 2).Run(ctx, []event.Event{newTestEvent(typeA)})
		if !errors.Is(err, ErrBFSDepthExceeded) {
			t.Fatalf("expected ErrBFSDepthExceeded, got %v", err)
		}
	})

	t.Run("EventLoopDetected", func(t *testing.T) {
		dupUUID := "dup-uuid-123"
		evt1 := event.Event{Uuid: dupUUID, EventType: typeA, AggregateUuid: "agg-1"}
		evt2 := event.Event{Uuid: dupUUID, EventType: typeA, AggregateUuid: "agg-1"}
		exec := NewExecutor().Register(typeA, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
			return HandlerResult{}, nil
		}))
		err := NewBFSEngine(exec, 10).Run(ctx, []event.Event{evt1, evt2})
		if !errors.Is(err, ErrEventLoopDetected) {
			t.Fatalf("expected ErrEventLoopDetected, got %v", err)
		}
	})

	t.Run("HandlerNotFound", func(t *testing.T) {
		exec := NewExecutor()
		err := NewBFSEngine(exec, 10).Run(ctx, []event.Event{newTestEvent(typeA)})
		if !errors.Is(err, ErrHandlerNotFound) {
			t.Fatalf("expected ErrHandlerNotFound, got %v", err)
		}
	})

	t.Run("HandlerError_Propagated", func(t *testing.T) {
		customErr := fmt.Errorf("domain error")
		exec := NewExecutor().Register(typeA, HandlerFunc(func(_ context.Context, _ event.Event) (HandlerResult, error) {
			return HandlerResult{}, customErr
		}))
		err := NewBFSEngine(exec, 10).Run(ctx, []event.Event{newTestEvent(typeA)})
		if !errors.Is(err, customErr) {
			t.Fatalf("expected customErr, got %v", err)
		}
	})
}
