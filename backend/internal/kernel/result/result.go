package result

import "akatengu/internal/kernel/event"

type EventResult struct {
	Events []event.Event
	State  any // optional aggregate state produced by the handler, consumed by projectors
}
