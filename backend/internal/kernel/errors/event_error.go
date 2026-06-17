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
	Retryable     bool
	Fatal         bool
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
	err := NewEventError(code, CategoryContract, evt, cause)
	err.Retryable = false
	err.Fatal = true
	return err
}

func NewRuntimeError(
	code EventErrorCode,
	evt event.Event,
	cause error,
	retryable bool,
	fatal bool,
) EventError {
	err := NewEventError(code, CategoryRuntime, evt, cause)
	err.Retryable = retryable
	err.Fatal = fatal
	return err
}

func NewBusinessError(
	code EventErrorCode,
	evt event.Event,
	cause error,
) EventError {
	err := NewEventError(code, CategoryBusiness, evt, cause)
	err.Retryable = false
	err.Fatal = false
	return err
}

func NewInfrastructureError(
	code EventErrorCode,
	evt event.Event,
	cause error,
) EventError {
	err := NewEventError(code, CategoryInfrastructure, evt, cause)
	err.Retryable = true
	err.Fatal = false
	return err
}
