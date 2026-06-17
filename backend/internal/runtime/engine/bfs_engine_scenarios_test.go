package engine

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/runtime/mediator"
	"context"
	"errors"
)

var errCustomHandler = errors.New("custom handler error")

type bfsScenario struct {
	given    string
	when     string
	then     string
	events   func() []event.Event
	executor func() *Executor
	maxDepth int
	wantErr  error
}

var bfsScenarios = []bfsScenario{
	{
		given:    "event type 未在 Executor 登記",
		when:     "Run 被呼叫",
		then:     "回傳 ErrHandlerNotFound",
		maxDepth: 10,
		events:   func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor { return NewExecutor(mediator.NewMediator()) },
		wantErr:  mediator.ErrHandlerNotFound,
	},
	{
		given:    "handler 每次都回傳子事件，maxDepth=2",
		when:     "Run 被呼叫",
		then:     "depth 超過上限，回傳 ErrBFSDepthExceeded",
		maxDepth: 2,
		events:   func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
				return mediator.HandlerResult{Children: []event.Event{makeTestEvent(typeA)}}, nil
			}))
			return NewExecutor(med)
		},
		wantErr: ErrBFSDepthExceeded,
	},
	{
		given:    "初始 batch 中包含兩個相同 UUID 的 event",
		when:     "Run 被呼叫",
		then:     "偵測到迴圈，回傳 ErrEventLoopDetected",
		maxDepth: 10,
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
		wantErr: ErrEventLoopDetected,
	},
	{
		given:    "handler 回傳錯誤",
		when:     "Run 被呼叫",
		then:     "handler 的錯誤原封不動從 Run 回傳",
		maxDepth: 10,
		events:   func() []event.Event { return []event.Event{makeTestEvent(typeA)} },
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ event.Event) (mediator.HandlerResult, error) {
				return mediator.HandlerResult{}, errCustomHandler
			}))
			return NewExecutor(med)
		},
		wantErr: errCustomHandler,
	},
}
