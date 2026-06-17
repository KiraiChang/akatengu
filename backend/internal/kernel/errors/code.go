package errors

type EventErrorCode string

// ========== CONTRACT LAYER ==========
const (
	ErrUnknownEventType   EventErrorCode = "EVENT_UNKNOWN_TYPE"
	ErrUnsupportedVersion EventErrorCode = "EVENT_UNSUPPORTED_VERSION"
	ErrDecodePayload      EventErrorCode = "EVENT_DECODE_PAYLOAD"
	ErrSchemaMismatch     EventErrorCode = "EVENT_SCHEMA_MISMATCH"
)

// ========== FLOW / RUNTIME LAYER ==========
const (
	ErrHandlerNotFound   EventErrorCode = "RUNTIME_HANDLER_NOT_FOUND"
	ErrExecutionTimeout  EventErrorCode = "RUNTIME_EXECUTION_TIMEOUT"
	ErrBFSDepthExceeded  EventErrorCode = "RUNTIME_BFS_DEPTH_EXCEEDED"
	ErrEventLoopDetected EventErrorCode = "RUNTIME_EVENT_LOOP_DETECTED"
)

// ========== BUSINESS LAYER ==========
const (
	ErrBusinessRuleFailed EventErrorCode = "BUSINESS_RULE_FAILED"
	ErrInvariantViolation EventErrorCode = "BUSINESS_INVARIANT_VIOLATION"
)

// ========== INFRASTRUCTURE LAYER ==========
const (
	ErrStoreWriteFailed   EventErrorCode = "INFRA_EVENTSTORE_WRITE_FAILED"
	ErrStoreReadFailed    EventErrorCode = "INFRA_EVENTSTORE_READ_FAILED"
	ErrQueueOverflow      EventErrorCode = "INFRA_QUEUE_OVERFLOW"
	ErrBackpressureActive EventErrorCode = "INFRA_BACKPRESSURE_ACTIVE"
)
