package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
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

// ─── Spec runner ─────────────────────────────────────────────────────────────

var _ = Describe("BFSEngine Run", func() {
	ctx := context.Background()

	// --- 錯誤情境（scenario table）---
	for _, s := range bfsScenarios {
		s := s
		label := fmt.Sprintf("GIVEN %s\n  WHEN %s\n  THEN %s", s.given, s.when, s.then)
		It(label, func() {
			eng := NewBFSEngine(s.executor(), s.maxDepth)
			err := eng.Run(ctx, s.events())
			ee, ok := extractEventError(err)
			Expect(ok).To(BeTrue(), fmt.Sprintf("expected EventError, got %T: %v", err, err))
			Expect(ee.Code).To(Equal(s.wantCode))
		})
	}

	// --- 執行驗證情境 ---

	Context("GIVEN 一個 event，handler 回傳空 Children", func() {
		When("Run 被呼叫", func() {
			It("THEN handler 被執行且回傳 nil", func() {
				called := false
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
					called = true
					return mediator.HandlerResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med), 10)
				Expect(eng.Run(ctx, []event.Event{makeTestEvent(typeA)})).To(BeNil())
				Expect(called).To(BeTrue())
			})
		})
	})

	Context("GIVEN root event 發射一個 typeB child", func() {
		When("Run 被呼叫", func() {
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
				eng := NewBFSEngine(NewExecutor(med), 10)
				Expect(eng.Run(ctx, []event.Event{makeTestEvent(typeA)})).To(BeNil())
				Expect(order).To(Equal([]string{"A", "B"}))
			})
		})
	})
})
