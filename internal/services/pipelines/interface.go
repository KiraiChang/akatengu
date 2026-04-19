package pipelines

import (
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"context"
)

// ------------------------------
// interface
// ------------------------------

type Projector[S any, P payload.Payload] interface {
	Project(ctx context.Context, ct *Context[S, P]) error
}

type PipelineRegistry interface {
	Dispatch(ctx context.Context, cmd cmd.AppendCmd) (*Result, error)
}

type PipelineRunner interface {
	Run(ctx context.Context, cmd cmd.AppendCmd) (*Result, error)
}
