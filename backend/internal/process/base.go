package process

import (
	"akatengu/internal/kernel/errors"
	"akatengu/internal/kernel/event"
	"fmt"
)

func checkAndGetPayload[T any](evt event.Event) (*T, error) {
	p, ok := evt.Payload.(T)
	if !ok {
		return nil, errors.NewBusinessError(errors.ErrPayloadInvalid, evt, fmt.Errorf("payload type error"))
	}
	return &p, nil
}
