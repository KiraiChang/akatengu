package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"context"
	"time"
)

type executorScenario struct {
	given, when, then string
	batch             func() []event.Event
	setup             func() *Executor
	wantUUIDs         []string
	wantCode          kerrors.EventErrorCode
}

var executorScenarios = []executorScenario{
	{
		given: "單一 event，handler 回傳空 children",
		when:  "ExecuteBatch 被呼叫",
		then:  "回傳空 children",
		batch: func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		setup: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
				return result.EventResult{}, nil
			}))
			return NewExecutor(med)
		},
		wantUUIDs: []string{},
	},
	{
		given: "單一 event，handler 回傳兩個 children",
		when:  "ExecuteBatch 被呼叫",
		then:  "回傳兩個 children，UUID 依序為 c1、c2",
		batch: func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		setup: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, _ event.Event) (result.EventResult, error) {
				return result.EventResult{Events: []event.Event{
					{Uuid: "c1", EventType: typeB},
					{Uuid: "c2", EventType: typeB},
				}}, nil
			}))
			return NewExecutor(med)
		},
		wantUUIDs: []string{"c1", "c2"},
	},
	{
		given: "handler 回傳 EventError",
		when:  "ExecuteBatch 被呼叫",
		then:  "error 被傳遞，Code 為 ErrBusinessRuleFailed",
		batch: func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		setup: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
				return result.EventResult{}, kerrors.NewBusinessError(kerrors.ErrBusinessRuleFailed, evt, nil)
			}))
			return NewExecutor(med)
		},
		wantCode: kerrors.ErrBusinessRuleFailed,
	},
	{
		// 驗證 childrenByIdx 的排序保證：B goroutine 先完成，A goroutine 後完成，
		// children 仍依 batch 插入順序（A 在前、B 在後）輸出。
		// 若改用 naive concurrent append，B 先完成會得到 [E, C, D]，測試紅掉。
		given: "batch [A, B]，A 的 handler 需 20ms，B 的 handler 立即完成",
		when:  "ExecuteBatch 非同步分派",
		then:  "children 仍依 A→B 插入順序排列為 [C, D, E]，不受 goroutine 完成時序影響",
		batch: func() []event.Event {
			return []event.Event{
				{Uuid: "A", EventType: typeA},
				{Uuid: "B", EventType: typeA},
			}
		},
		setup: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
				switch evt.Uuid {
				case "A":
					time.Sleep(20 * time.Millisecond) // 慢，確保 B goroutine 先完成
					return result.EventResult{Events: []event.Event{
						{Uuid: "C", EventType: typeB},
						{Uuid: "D", EventType: typeB},
					}}, nil
				case "B":
					return result.EventResult{Events: []event.Event{
						{Uuid: "E", EventType: typeB},
					}}, nil
				}
				return result.EventResult{}, nil
			}))
			return NewExecutor(med)
		},
		wantUUIDs: []string{"C", "D", "E"},
	},
}
