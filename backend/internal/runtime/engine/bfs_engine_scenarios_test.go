package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
)

type bfsScenario struct {
	given       string
	when        string
	then        string
	events      func() []event.Event
	executor    func() *Executor
	middlewares []safety.Middleware
	wantCode    kerrors.EventErrorCode
}

var bfsScenarios = []bfsScenario{
	{
		given:    "event type 未在 Executor 登記",
		when:     "Run 被呼叫",
		then:     "回傳 EventError，Code 為 ErrHandlerNotFound",
		events:   func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor { return NewExecutor(mediator.NewMediator()) },
		wantCode: kerrors.ErrHandlerNotFound,
	},
	{
		given:  "handler 每次都回傳子事件，DepthGuard maxDepth=2",
		when:   "Run 被呼叫",
		then:   "回傳 EventError，Code 為 ErrBFSDepthExceeded",
		events: func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
				return mediator.HandlerResult{Children: []event.Event{makeTestEvent(typeA)}}, nil
			}))
			return NewExecutor(med)
		},
		middlewares: []safety.Middleware{safety.DepthGuard(2)},
		wantCode:    kerrors.ErrBFSDepthExceeded,
	},
	{
		given: "初始 batch 中包含兩個相同 UUID 的 event，CycleDetector 啟用",
		when:  "Run 被呼叫",
		then:  "回傳 EventError，Code 為 ErrEventLoopDetected",
		events: func() []event.Event {
			dup := event.Event{Uuid: "dup-uuid-abc", EventType: typeA, AggregateUuid: "agg-1"}
			return []event.Event{dup, dup}
		},
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
				return mediator.HandlerResult{}, nil
			}))
			return NewExecutor(med)
		},
		middlewares: []safety.Middleware{safety.CycleDetector()},
		wantCode:    kerrors.ErrEventLoopDetected,
	},
	{
		given:  "handler 回傳業務規則 EventError",
		when:   "Run 被呼叫",
		then:   "EventError 原封不動從 Run 回傳，Code 為 ErrBusinessRuleFailed",
		events: func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, evt event.Event) (mediator.HandlerResult, error) {
				return mediator.HandlerResult{}, kerrors.NewBusinessError(kerrors.ErrBusinessRuleFailed, evt, nil)
			}))
			return NewExecutor(med)
		},
		wantCode: kerrors.ErrBusinessRuleFailed,
	},
}
