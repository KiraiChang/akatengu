package engine

import (
	kerrors "akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"context"
)

type bfsScenario struct {
	given       string
	when        string
	then        string
	evt         func() event.Event
	executor    func() *Executor
	middlewares []safety.Middleware
	wantCode    kerrors.EventErrorCode
}

var bfsScenarios = []bfsScenario{
	{
		given:    "event type 未在 Executor 登記",
		when:     "Run 被呼叫",
		then:     "回傳 EventError，Code 為 ErrHandlerNotFound",
		evt:      func() event.Event { return makeTestEvent(typeA) },
		executor: func() *Executor { return NewExecutor(mediator.NewMediator()) },
		wantCode: kerrors.ErrHandlerNotFound,
	},
	{
		given: "handler 回傳業務規則 EventError",
		when:  "Run 被呼叫",
		then:  "EventError 原封不動從 Run 回傳，Code 為 ErrBusinessRuleFailed",
		evt:   func() event.Event { return makeTestEvent(typeA) },
		executor: func() *Executor {
			med := mediator.NewMediator()
			med.Register(typeA, mediator.HandlerFunc(func(_ context.Context, _ *query.Repo, evt event.Event) (result.EventResult, error) {
				return result.EventResult{}, kerrors.NewBusinessError(kerrors.ErrBusinessRuleFailed, evt, nil)
			}))
			return NewExecutor(med)
		},
		wantCode: kerrors.ErrBusinessRuleFailed,
	},
}
