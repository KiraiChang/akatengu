package errors

import (
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"fmt"
)

type EventError struct {
	Code          EventErrorCode
	Category      ErrorCategory
	EventType     event_types.EventType
	Version       int
	AggregateUuid string
	Cause         error
}

func (e EventError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s", e.Category, e.Code, e.Cause.Error())
	}
	return fmt.Sprintf("[%s] %s", e.Category, e.Code)
}

func NewEventError(
	code EventErrorCode,
	category ErrorCategory,
	evt event.Event,
	cause error,
) EventError {
	return EventError{
		Code:          code,
		Category:      category,
		EventType:     evt.EventType,
		Version:       evt.Version,
		AggregateUuid: evt.AggregateUuid,
		Cause:         cause,
	}
}

func NewContractError(
	code EventErrorCode,
	evt event.Event,
	cause error,
) EventError {
	return NewEventError(code, CategoryContract, evt, cause)
}

// NewRuntimeError creates a runtime EventError. cause is optional; pass nil when absent.
// Retry/fatal policy is now expressed via result.Policy returned by EventProcessor,
// not embedded in the error itself.
func NewRuntimeError(
	code EventErrorCode,
	evt event.Event,
	cause ...error,
) EventError {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return NewEventError(code, CategoryRuntime, evt, c)
}

func NewBusinessError(
	code EventErrorCode,
	evt event.Event,
	cause error,
) EventError {
	return NewEventError(code, CategoryBusiness, evt, cause)
}

func NewInfrastructureError(
	code EventErrorCode,
	evt event.Event,
	cause error,
) EventError {
	return NewEventError(code, CategoryInfrastructure, evt, cause)
}
