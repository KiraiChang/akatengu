package engine

import (
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

type testPayload struct{ typ event_types.EventType }

func (p testPayload) EventType() event_types.EventType { return p.typ }

func makeTestEvent(typ event_types.EventType) event.Event {
	return event.NewEvent[testPayload]("agg-1", testPayload{typ: typ})
}

func extractEventError(err error) (kerrors.EventError, bool) {
	var ee kerrors.EventError
	return ee, errors.As(err, &ee)
}

func defaultMiddlewares() []safety.Middleware {
	return []safety.Middleware{safety.DepthGuard(10), safety.CycleDetector()}
}

// ─── Spec runner ─────────────────────────────────────────────────────────────

var _ = Describe("BFSEngine Run", func() {
	ctx := context.Background()

	// --- 錯誤情境（scenario table）---
	for _, s := range bfsScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)
		It(label, func() {
			eng := NewBFSEngine(s.executor()).WithMiddleware(s.middlewares...)
			err := eng.Run(ctx, s.evt())
			ee, ok := extractEventError(err)
			Expect(ok).To(BeTrue(), fmt.Sprintf("expected EventError, got %T: %v", err, err))
			Expect(ee.Code).To(Equal(s.wantCode))
		})
	}

	// --- 執行驗證情境 ---

	Context("GIVEN 一個 event，handler 回傳空 Children", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			It("THEN handler 被執行且回傳 nil", func() {
				called := false
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (result.EventResult, error) {
					called = true
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				Expect(eng.Run(ctx, makeTestEvent(typeA))).To(BeNil())
				Expect(called).To(BeTrue())
			})
		})
	})

	Context("GIVEN root event 發射一個 typeB child", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			It("THEN 兩個 handler 依 BFS 順序各被執行一次，回傳 nil", func() {
				var order []string
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (result.EventResult, error) {
					order = append(order, "A")
					return result.EventResult{Events: []event.Event{makeTestEvent(typeB)}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (result.EventResult, error) {
					order = append(order, "B")
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				Expect(eng.Run(ctx, makeTestEvent(typeA))).To(BeNil())
				Expect(order).To(Equal([]string{"A", "B"}))
			})
		})
	})

	Context("GIVEN A 發射 [B, C]，B 發射 [D]，C 發射 [E]", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			// FIFO Queue 保證 BFS level 順序：depth=0 的 A 先完成後，
			// depth=1 的 B、C 依插入順序依序處理，depth=2 的 D、E 最後。
			It("THEN 執行順序為 BFS level 順序 [A, B, C, D, E]", func() {
				var order []string
				evtB := event.Event{Uuid: "B", EventType: typeA}
				evtC := event.Event{Uuid: "C", EventType: typeA}
				evtD := event.Event{Uuid: "D", EventType: typeB}
				evtE := event.Event{Uuid: "E", EventType: typeB}

				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, evt event.Event) (result.EventResult, error) {
					order = append(order, evt.Uuid)
					switch evt.Uuid {
					case "B":
						return result.EventResult{Events: []event.Event{evtD}}, nil
					case "C":
						return result.EventResult{Events: []event.Event{evtE}}, nil
					}
					return result.EventResult{Events: []event.Event{evtB, evtC}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, evt event.Event) (result.EventResult, error) {
					order = append(order, evt.Uuid)
					return result.EventResult{}, nil
				}))

				evtA := event.Event{Uuid: "A", EventType: typeA}
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				Expect(eng.Run(ctx, evtA)).To(BeNil())
				Expect(order).To(Equal([]string{"A", "B", "C", "D", "E"}))
			})
		})
	})
})
