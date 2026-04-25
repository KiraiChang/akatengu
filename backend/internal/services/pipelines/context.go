package pipelines

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/payload"
)

type NoState struct{}

type Context[S any, P payload.Payload] struct {
	Payload P
	State   *S
}

type Result struct {
	Payload payload.Payload
	State   any
	Event   db.EventStore
}
