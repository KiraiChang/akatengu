package pipelines

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"context"
	"encoding/json"
)

// ------------------------------
// Pipeline
// ------------------------------

type Pipeline[S any, P payload.Payload] struct {
	projector Projector[S, P]
}

func (p *Pipeline[S, P]) Run(ctx context.Context, ct *Context[S, P]) error {
	return p.projector.Project(ctx, ct)
}

// ------------------------------
// TypedPipeline
// ------------------------------

type TypedPipeline[S any, P payload.Payload] struct {
	pipeline *Pipeline[S, P]
	factory  func() *S
}

func NewType[S any, P payload.Payload](projector Projector[S, P], f func() *S) *TypedPipeline[S, P] {
	pipeline := &Pipeline[S, P]{
		projector: projector,
	}

	return &TypedPipeline[S, P]{
		pipeline: pipeline,
		factory:  f,
	}
}

func NewTypeWithNoState[P payload.Payload](projector Projector[NoState, P]) *TypedPipeline[NoState, P] {
	return NewType[NoState, P](projector, func() *NoState {
		return &NoState{}
	})
}

func (t *TypedPipeline[S, P]) Run(ctx context.Context, cmd cmd.AppendCmd) (*Result, error) {
	// 1. 型別轉換
	var p P
	if err := json.Unmarshal(cmd.Payload, &p); err != nil {
		return nil, err
	}

	// 2. 建立 context（保證不會 nil）
	ct := &Context[S, P]{
		State:   t.factory(),
		Payload: p,
	}

	// 3. 執行
	if err := t.pipeline.Run(ctx, ct); err != nil {
		return nil, err
	}

	// 4. 結果
	return &Result{
		Payload: p,
		State:   ct.State,
	}, nil
}
