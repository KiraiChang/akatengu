package result

import "akatengu/internal/kernel/event"

// Policy describes how the BFS engine should react when event processing fails.
// Middlewares set a default policy for their own errors; handlers may override
// by returning a non-zero Policy alongside an error.
type Policy struct {
	Retryable bool
	Fatal     bool
}

type EventResult struct {
	Events []event.Event
	// Payload, when set, replaces the payload of the event the handler handled
	// before the event is written, so the event records what the handler
	// decided (ADR-023 principle 1, BFS R3 D3). It must have the same type.
	Payload any
	Policy  Policy // error-handling policy; zero value = not retryable, not fatal
}
