package handle

import (
	"akatengu/internal/kernel/event"
	"akatengu/internal/services/pipelines"
	"fmt"
)

func toUpdatedBy(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func checkAndGetPayload[T any](ct *event.Event) (*T, error) {
	p, ok := ct.Payload.(T)
	if !ok {
		return nil, fmt.Errorf("payload type error")
	}
	return &p, nil
}

func checkAndGetState[T any](ct *pipelines.Result) (*T, error) {
	p, ok := ct.State.(*T)
	if !ok {
		return nil, fmt.Errorf("state type error")
	}
	return p, nil
}
