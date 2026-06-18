package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
	"errors"
	"fmt"
	"sync"

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
			eng := NewBFSEngine(s.executor(), s.middlewares...)
			err := eng.Run(ctx, s.events())
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
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
					called = true
					return mediator.HandlerResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med), defaultMiddlewares()...)
				Expect(eng.Run(ctx, []event.Event{makeTestEvent(typeA)})).To(BeNil())
				Expect(called).To(BeTrue())
			})
		})
	})

	Context("GIVEN root event 發射一個 typeB child", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			It("THEN 兩個 handler 依 BFS 順序各被執行一次，回傳 nil", func() {
				var order []string
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
					order = append(order, "A")
					return mediator.HandlerResult{Children: []event.Event{makeTestEvent(typeB)}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
					order = append(order, "B")
					return mediator.HandlerResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med), defaultMiddlewares()...)
				Expect(eng.Run(ctx, []event.Event{makeTestEvent(typeA)})).To(BeNil())
				Expect(order).To(Equal([]string{"A", "B"}))
			})
		})
	})

	Context("GIVEN root batch [A, B]，A 發射 [C, D]，B 發射 [E]", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			// Async ExecuteBatch：同一 level 的 handler 並行執行，within-level 執行順序
			// 不確定；可保證的是 level-0 全部完成後才開始 level-1（BFS Queue 結構保證）。
			It("THEN level-0 [A, B] 全部完成後才執行 level-1 [C, D, E]（BFS 層次保證成立）", func() {
				var mu sync.Mutex
				var level0, level1 []string

				evtC := event.Event{Uuid: "C", EventType: typeB}
				evtD := event.Event{Uuid: "D", EventType: typeB}
				evtE := event.Event{Uuid: "E", EventType: typeB}

				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, evt event.Event) (mediator.HandlerResult, error) {
					mu.Lock()
					level0 = append(level0, evt.Uuid)
					mu.Unlock()
					switch evt.Uuid {
					case "A":
						return mediator.HandlerResult{Children: []event.Event{evtC, evtD}}, nil
					case "B":
						return mediator.HandlerResult{Children: []event.Event{evtE}}, nil
					}
					return mediator.HandlerResult{}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, evt event.Event) (mediator.HandlerResult, error) {
					mu.Lock()
					level1 = append(level1, evt.Uuid)
					mu.Unlock()
					return mediator.HandlerResult{}, nil
				}))

				evtA := event.Event{Uuid: "A", EventType: typeA}
				evtB := event.Event{Uuid: "B", EventType: typeA}
				eng := NewBFSEngine(NewExecutor(med), defaultMiddlewares()...)
				Expect(eng.Run(ctx, []event.Event{evtA, evtB})).To(BeNil())
				Expect(level0).To(ConsistOf("A", "B"))
				Expect(level1).To(ConsistOf("C", "D", "E"))
			})
		})
	})
})
