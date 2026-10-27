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
	ErrProjectorNotFound EventErrorCode = "RUNTIME_PROJECTOR_NOT_FOUND"
	ErrProjectorError    EventErrorCode = "RUNTIME_PROJECTOR_ERROR"
	ErrCheckPointError   EventErrorCode = "RUNTIME_CHECKPOINT_ERROR"
)

// ErrChildEnvelopeInvalid: a child event's aggregate cannot be accepted by
// BFSEngine (ADR-023 A2).
const ErrChildEnvelopeInvalid EventErrorCode = "RUNTIME_CHILD_ENVELOPE_INVALID"

// ErrRootEnvelopeInvalid: a root event's version mode is unknown to BFSEngine
// (ADR-027 rule 4).
const ErrRootEnvelopeInvalid EventErrorCode = "RUNTIME_ROOT_ENVELOPE_INVALID"

// ErrResultPayloadInvalid: a handler returned a payload whose type is not the
// one of the event it handled (BFS R3 D3).
const ErrResultPayloadInvalid EventErrorCode = "RUNTIME_RESULT_PAYLOAD_INVALID"

// ========== BUSINESS LAYER ==========
const (
	ErrBusinessRuleFailed EventErrorCode = "BUSINESS_RULE_FAILED"
	ErrPayloadInvalid     EventErrorCode = "BUSINESS_PAYLOAD_INVALID"
	ErrInvariantViolation EventErrorCode = "BUSINESS_INVARIANT_VIOLATION"
)

// ========== INFRASTRUCTURE LAYER ==========
const (
	ErrStoreWriteFailed   EventErrorCode = "INFRA_EVENTSTORE_WRITE_FAILED"
	ErrStoreReadFailed    EventErrorCode = "INFRA_EVENTSTORE_READ_FAILED"
	ErrVersionConflict    EventErrorCode = "INFRA_VERSION_CONFLICT"
	ErrQueueOverflow      EventErrorCode = "INFRA_QUEUE_OVERFLOW"
	ErrBackpressureActive EventErrorCode = "INFRA_BACKPRESSURE_ACTIVE"
)
