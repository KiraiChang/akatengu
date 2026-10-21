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
	Policy Policy // error-handling policy; zero value = not retryable, not fatal
}
