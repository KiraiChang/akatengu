package pipelines

import "akatengu/internal/model/db"

type Payload interface {
}

type NoState struct{}

type Context[S any, P Payload] struct {
	Payload P
	State   *S
}

type Result struct {
	Payload Payload
	State   any
	Event   db.EventStore
}
