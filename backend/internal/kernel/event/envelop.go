package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"time"

	"github.com/google/uuid"
)

type Aggregate struct {
	ID string
}

// Payload 讓payload去定義的interface
type Payload interface {
	EventType() event_types.EventType
}

// Event 事件的封套
type Event struct {
	Uuid          string
	AggregateUuid string
	AggregateType enums.AggregateType
	Version       int
	EventType     event_types.EventType
	Metadata      Metadata
	Payload       any
}

func NewEvent[T Payload](
	aggID string,
	payload T,
	meta ...Metadata,
) Event {
	m := Metadata{}

	if len(meta) > 0 {
		m = meta[0]
	}

	if m.Tracing.CorrelationID == "" {
		m.Tracing.CorrelationID = uuid.NewString()
	}

	return Event{
		Uuid:          uuid.NewString(),
		AggregateUuid: aggID,
		EventType:     payload.EventType(),
		Metadata:      m,
		Payload:       payload,
	}
}

func NewNextEvent[T Payload](parent Event, payload T) Event {
	return Event{
		Uuid:          uuid.NewString(),
		AggregateUuid: parent.AggregateUuid,
		EventType:     payload.EventType(),
		Payload:       payload,
		Metadata: Metadata{
			Tracing: TracingData{
				CorrelationID: parent.Metadata.Tracing.CorrelationID,
				CausationID:   parent.Uuid,
			},
			Execution: ExecutionData{
				PartitionKey: parent.Metadata.Execution.PartitionKey,
				Source:       "process",
			},
			Schema: parent.Metadata.Schema,
			Observability: ObservabilityData{
				Host:      parent.Metadata.Observability.Host,
				Timestamp: time.Now().Unix(),
			},
		},
	}
}
