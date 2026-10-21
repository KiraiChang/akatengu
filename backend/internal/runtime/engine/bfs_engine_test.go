package engine

import (
	"akatengu/internal/enums/event_types"
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

func noopHandler() mediator.HandlerFunc {
	return func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
		return result.EventResult{}, nil
	}
}

// emitChildren returns n typeB children of parent with UUIDs unique to parent.
func emitChildren(parent event.Event, n int) []event.Event {
	children := make([]event.Event, n)
	for i := range children {
		children[i] = event.Event{Uuid: fmt.Sprintf("%s-child-%d", parent.Uuid, i), EventType: typeB}
	}
	return children
}

func causedBy(uuid string, typ event_types.EventType, causationID string) event.Event {
	return event.Event{Uuid: uuid, EventType: typ, Metadata: event.Metadata{Tracing: event.TracingData{CausationID: causationID}}}
}

// barrier releases its callers once n of them have arrived, forcing concurrent
// Runs to overlap inside their handlers.
type barrier struct {
	n       int
	mu      sync.Mutex
	arrived int
	release chan struct{}
}

func newBarrier(n int) *barrier {
	return &barrier{n: n, release: make(chan struct{})}
}

func (b *barrier) wait() error {
	b.mu.Lock()
	b.arrived++
	if b.arrived == b.n {
		close(b.release)
	}
	b.mu.Unlock()
	return waitClosed(b.release)
}

// waitClosed waits for ch to be closed, failing instead of hanging the suite.
func waitClosed(ch <-chan struct{}) error {
	select {
	case <-ch:
		return nil
	case <-time.After(5 * time.Second):
		return errors.New("timed out waiting for the concurrent Run")
	}
}

// runConcurrently runs each root on eng in its own goroutine and returns the
// errors in root order.
func runConcurrently(ctx context.Context, eng *BFSEngine, roots ...event.Event) []error {
	errs := make([]error, len(roots))
	var wg sync.WaitGroup
	for i, root := range roots {
		wg.Add(1)
		go func(i int, root event.Event) {
			defer GinkgoRecover()
			defer wg.Done()
			_, errs[i] = eng.Run(ctx, root)
		}(i, root)
	}
	wg.Wait()
	return errs
}

func expectEventErrorCode(err error, code kerrors.EventErrorCode) {
	ee, ok := extractEventError(err)
	Expect(ok).To(BeTrue(), fmt.Sprintf("expected EventError, got %T: %v", err, err))
	Expect(ee.Code).To(Equal(code))
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
			_, err := eng.Run(ctx, s.evt())
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
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					called = true
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(BeNil())
				Expect(called).To(BeTrue())
			})
		})
	})

	Context("GIVEN root event 發射一個 typeB child", func() {
		When("Run 被呼叫（含 DepthGuard + CycleDetector）", func() {
			It("THEN 兩個 handler 依 BFS 順序各被執行一次，回傳 nil", func() {
				var order []string
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					order = append(order, "A")
					return result.EventResult{Events: []event.Event{makeTestEvent(typeB)}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					order = append(order, "B")
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(BeNil())
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
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					order = append(order, evt.Uuid)
					switch evt.Uuid {
					case "B":
						return result.EventResult{Events: []event.Event{evtD}}, nil
					case "C":
						return result.EventResult{Events: []event.Event{evtE}}, nil
					}
					return result.EventResult{Events: []event.Event{evtB, evtC}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					order = append(order, evt.Uuid)
					return result.EventResult{}, nil
				}))

				evtA := event.Event{Uuid: "A", EventType: typeA}
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				_, err := eng.Run(ctx, evtA)
				Expect(err).To(BeNil())
				Expect(order).To(Equal([]string{"A", "B", "C", "D", "E"}))
			})
		})
	})

	// ─── UnitOfWork 路徑 ──────────────────────────────────────────────────────

	Context("GIVEN UnitOfWork 已設定", func() {
		When("handler 成功（單一事件）", func() {
			It("THEN Store.Append 被呼叫 1 次", func() {
				store := &mockStore{}
				uow := &mockUoW{store: store}
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).
					WithMiddleware(defaultMiddlewares()...).
					WithUnitOfWork(uow)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(BeNil())
				Expect(store.appendCount).To(Equal(1))
			})
		})

		When("BFS A → B（兩個事件依序處理）", func() {
			It("THEN Store.Append 被呼叫 2 次", func() {
				store := &mockStore{}
				uow := &mockUoW{store: store}
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					return result.EventResult{Events: []event.Event{makeTestEvent(typeB)}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).
					WithMiddleware(defaultMiddlewares()...).
					WithUnitOfWork(uow)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(BeNil())
				Expect(store.appendCount).To(Equal(2))
			})
		})

		When("Store.Append 回傳錯誤", func() {
			It("THEN Run 回傳該錯誤", func() {
				storeErr := errors.New("store failed")
				store := &mockStore{err: storeErr}
				uow := &mockUoW{store: store}
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).
					WithMiddleware(defaultMiddlewares()...).
					WithUnitOfWork(uow)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(MatchError(storeErr))
			})
		})
	})

	Context("GIVEN UnitOfWork + Projector 已設定", func() {
		When("handler 成功，projector 型別符合事件", func() {
			It("THEN Projector.Apply 被呼叫 1 次", func() {
				store := &mockStore{}
				uow := &mockUoW{store: store}
				proj := &mockProjector{types: []event_types.EventType{typeA}}
				registry := handle.NewRegistry()
				registry.Register(proj)
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).
					WithMiddleware(defaultMiddlewares()...).
					WithUnitOfWork(uow).
					WithProjector(registry)
				_, err := eng.Run(ctx, makeTestEvent(typeA))
				Expect(err).To(BeNil())
				Expect(proj.applyCount).To(Equal(1))
			})
		})
	})

	// ─── middleware 狀態隔離（ISSUE-025）─────────────────────────────────────

	Context("GIVEN 同一個 engine 依序執行多次 Run（含 DepthGuard + CycleDetector）", func() {
		When("同一個 root event 再次 Run", func() {
			It("THEN 兩次都成功，不會被誤判為迴圈", func() {
				med := mediator.NewMediator()
				med.Register(typeA, noopHandler())
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				evt := event.Event{Uuid: "retry-root", EventType: typeA}

				_, err := eng.Run(ctx, evt)
				Expect(err).To(BeNil())
				_, err = eng.Run(ctx, evt)
				Expect(err).To(BeNil())
			})
		})

		When("同一次 Run 中出現重複的 UUID", func() {
			It("THEN 回傳 ErrEventLoopDetected", func() {
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					return result.EventResult{Events: []event.Event{{Uuid: evt.Uuid, EventType: typeB}}}, nil
				}))
				med.Register(typeB, noopHandler())
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)

				_, err := eng.Run(ctx, event.Event{Uuid: "dup", EventType: typeA})
				expectEventErrorCode(err, kerrors.ErrEventLoopDetected)
			})
		})

		When("第一次 Run 在子事件失敗，之後以同一個 root event 重試", func() {
			It("THEN 重試成功，不殘留前一次 Run 的狀態", func() {
				childErr := errors.New("child failed")
				failChild := true
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					return result.EventResult{Events: emitChildren(evt, 1)}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
					if failChild {
						failChild = false
						return result.EventResult{}, childErr
					}
					return result.EventResult{}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(defaultMiddlewares()...)
				evt := event.Event{Uuid: "failed-root", EventType: typeA}

				_, err := eng.Run(ctx, evt)
				Expect(err).To(MatchError(childErr))
				_, err = eng.Run(ctx, evt)
				Expect(err).To(BeNil())
			})
		})
	})

	Context("GIVEN 多個 goroutine 共用同一個 engine（NewMux 的 middleware 組態，不接 DB）", func() {
		When("以同步屏障強制各次 Run 在 handler 內重疊", func() {
			It("THEN 每次 Run 都成功，程式不會終止", func() {
				const runs = 8
				bar := newBarrier(runs)
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					if err := bar.wait(); err != nil {
						return result.EventResult{}, err
					}
					return result.EventResult{Events: emitChildren(evt, 2)}, nil
				}))
				med.Register(typeB, noopHandler())
				eng := NewBFSEngine(NewExecutor(med)).
					WithMiddleware(safety.BackpressureScheduler(3), safety.DepthGuard(10), safety.CycleDetector())

				roots := make([]event.Event, runs)
				for i := range roots {
					roots[i] = event.Event{Uuid: fmt.Sprintf("concurrent-root-%d", i), EventType: typeA}
				}
				for i, err := range runConcurrently(ctx, eng, roots...) {
					Expect(err).To(BeNil(), fmt.Sprintf("run %d", i))
				}
			})
		})
	})

	Context("GIVEN 兩個重疊的 Run 共用同一個 engine（BackpressureScheduler(3)）", func() {
		// The barrier makes both roots reset their counters first; the channels then
		// order the child accounting so a shared counter would deterministically overflow.
		When("兩個 root event 各發射 3 個子事件（各自未超量）", func() {
			It("THEN 兩次 Run 都成功", func() {
				bar := newBarrier(2)
				firstCounted := make(chan struct{})
				secondDone := make(chan struct{})
				var signalFirst sync.Once
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					if err := bar.wait(); err != nil {
						return result.EventResult{}, err
					}
					if evt.Uuid == "second-root" {
						// Account for the second root's children only after the first root's.
						if err := waitClosed(firstCounted); err != nil {
							return result.EventResult{}, err
						}
					}
					return result.EventResult{Events: emitChildren(evt, 3)}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					if evt.Uuid != "first-root-child-0" {
						return result.EventResult{}, nil
					}
					// Children run only after their root has been accounted for; keep the
					// first Run here until the second Run has finished.
					signalFirst.Do(func() { close(firstCounted) })
					return result.EventResult{}, waitClosed(secondDone)
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(safety.BackpressureScheduler(3))

				var firstErr, secondErr error
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					_, firstErr = eng.Run(ctx, event.Event{Uuid: "first-root", EventType: typeA})
				}()
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					defer close(secondDone)
					_, secondErr = eng.Run(ctx, event.Event{Uuid: "second-root", EventType: typeA})
				}()
				wg.Wait()

				Expect(firstErr).To(BeNil())
				Expect(secondErr).To(BeNil())
			})
		})

		When("一個 root event 發射 4 個子事件（超量），另一個發射 3 個", func() {
			It("THEN 超量的 Run 回傳 ErrQueueOverflow，另一個 Run 成功", func() {
				bar := newBarrier(2)
				overDone := make(chan struct{})
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					if err := bar.wait(); err != nil {
						return result.EventResult{}, err
					}
					if evt.Uuid == "over-root" {
						return result.EventResult{Events: emitChildren(evt, 4)}, nil
					}
					// Account for the in-limit root only after the overflowing Run has finished.
					if err := waitClosed(overDone); err != nil {
						return result.EventResult{}, err
					}
					return result.EventResult{Events: emitChildren(evt, 3)}, nil
				}))
				med.Register(typeB, noopHandler())
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(safety.BackpressureScheduler(3))

				var overErr, withinErr error
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					defer close(overDone)
					_, overErr = eng.Run(ctx, event.Event{Uuid: "over-root", EventType: typeA})
				}()
				go func() {
					defer GinkgoRecover()
					defer wg.Done()
					_, withinErr = eng.Run(ctx, event.Event{Uuid: "within-root", EventType: typeA})
				}()
				wg.Wait()

				expectEventErrorCode(overErr, kerrors.ErrQueueOverflow)
				Expect(withinErr).To(BeNil())
			})
		})
	})

	Context("GIVEN engine 使用 DeterministicLoopDetector", func() {
		When("同一次 Run 的因果鏈出現事件類型循環（A → B → A）", func() {
			It("THEN 回傳 ErrEventLoopDetected", func() {
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					return result.EventResult{Events: []event.Event{causedBy(evt.Uuid+"-b", typeB, evt.Uuid)}}, nil
				}))
				med.Register(typeB, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					return result.EventResult{Events: []event.Event{causedBy(evt.Uuid+"-a", typeA, evt.Uuid)}}, nil
				}))
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(safety.DeterministicLoopDetector())

				_, err := eng.Run(ctx, event.Event{Uuid: "loop-root", EventType: typeA})
				expectEventErrorCode(err, kerrors.ErrEventLoopDetected)
			})
		})

		When("第二次 Run 的 root event 由第一次 Run 的子事件引發", func() {
			It("THEN 第二次 Run 成功，不殘留第一次 Run 的因果圖", func() {
				med := mediator.NewMediator()
				med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
					return result.EventResult{Events: []event.Event{causedBy(evt.Uuid+"-b", typeB, evt.Uuid)}}, nil
				}))
				med.Register(typeB, noopHandler())
				eng := NewBFSEngine(NewExecutor(med)).WithMiddleware(safety.DeterministicLoopDetector())

				_, err := eng.Run(ctx, event.Event{Uuid: "first-root", EventType: typeA})
				Expect(err).To(BeNil())
				// With a leftover graph, first-root-b → first-root (typeA) would look like a loop.
				_, err = eng.Run(ctx, causedBy("second-root", typeA, "first-root-b"))
				Expect(err).To(BeNil())
			})
		})
	})
})
