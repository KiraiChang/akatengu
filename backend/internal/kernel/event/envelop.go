package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/pkg/uuidx"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AggregateID 以 aggType 為前綴、UUIDv7（時間序）為後綴的型別安全 ID
// 格式：{aggType}_{uuidv7}，例如 account_01956af1-3b2c-7f4d-a123-456789abcdef
type AggregateID string

func NewAggregateID(aggType enums.AggregateType) AggregateID {
	return AggregateID(fmt.Sprintf("%s_%s", aggType, uuid.Must(uuid.NewV7()).String()))
}

func ParseAggregateID(s string) AggregateID {
	return AggregateID(s)
}

func (a AggregateID) String() string {
	return string(a)
}

type Aggregate struct {
	ID AggregateID
}

func NewAggregate(aggType enums.AggregateType) Aggregate {
	return Aggregate{ID: NewAggregateID(aggType)}
}

// Payload 讓payload去定義的interface
type Payload interface {
	EventType() event_types.EventType
}

// ChildAggregate tells BFSEngine which aggregate a child event belongs to
// (ADR-023 A2, ADR-024). It has no meaning for a root event.
type ChildAggregate int

const (
	// ChildAggregateParent continues the parent's aggregate. It is the zero
	// value, so a child built without a constructor defaults to it.
	ChildAggregateParent ChildAggregate = iota
	// ChildAggregateDerived creates a new aggregate derived from the parent
	// event (NewDerivedEvent); it fails if the aggregate already exists.
	ChildAggregateDerived
)

// RootVersion tells BFSEngine how to version a root event, following the
// version policy of its event type (ADR-027 rule 4). It has no meaning for a
// child event, which follows ChildAggregate.
type RootVersion int

const (
	// RootVersionExpected writes the next version of an aggregate that must be
	// at ExpectedVersion; ExpectedVersion 0 creates it. It is the zero value, so
	// an event built without a mode keeps this behavior.
	RootVersionExpected RootVersion = iota
	// RootVersionNew creates the aggregate at version 1; any existing version
	// record, including one at version 0, is a conflict. ExpectedVersion is ignored.
	RootVersionNew
	// RootVersionCurrent writes the version after the one read inside the
	// transaction, creating the aggregate at version 1 when it has no record.
	// ExpectedVersion is ignored.
	RootVersionCurrent
)

// Event 事件的封套
type Event struct {
	Uuid            string                `json:"uuid"`
	AggregateUuid   string                `json:"aggregate_uuid"`
	AggregateType   enums.AggregateType   `json:"aggregate_type"`
	Version         int                   `json:"version"`
	EventType       event_types.EventType `json:"event_type"`
	Metadata        Metadata              `json:"metadata"`
	Payload         any                   `json:"payload"`
	UpdatedBy       string                `json:"update_by"`
	MerchantID      int64                 `json:"merchant_id"`
	ExpectedVersion int64                 `json:"expected_version"`
	Id              int64                 `json:"id"`
	ChildAggregate  ChildAggregate        `json:"-"`
	RootVersion     RootVersion           `json:"-"`
}

// EventParams 建立完整 Event 所需的外部輸入；Timestamp 自動填入，空值欄位套用預設值
type EventParams struct {
	RequestID     string
	CorrelationID string // 空則自動產生 UUIDv7
	CausationID   string
	PartitionKey  string // 空則 fallback 到 aggID
	Priority      int
	Source        string // "api" | "replay" | "process"
	Host          string
	SchemaVersion int
	Encoding      string // 空則預設 "json"
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
		m.Tracing.CorrelationID = NewUUIDv7()
	}

	return Event{
		Uuid:          NewUUIDv7(),
		AggregateUuid: aggID,
		EventType:     payload.EventType(),
		Metadata:      m,
		Payload:       payload,
	}
}

// NewFullEvent 建立帶有完整 Metadata 的 Event，自動填入 Timestamp 與各項預設值
func NewFullEvent[T Payload](aggID AggregateID, payload T, params EventParams) Event {
	if params.CorrelationID == "" {
		params.CorrelationID = NewUUIDv7()
	}
	if params.PartitionKey == "" {
		params.PartitionKey = aggID.String()
	}
	if params.Encoding == "" {
		params.Encoding = "json"
	}

	return Event{
		Uuid:          NewUUIDv7(),
		AggregateUuid: aggID.String(),
		EventType:     payload.EventType(),
		Payload:       payload,
		Metadata: Metadata{
			Tracing: TracingData{
				CorrelationID: params.CorrelationID,
				CausationID:   params.CausationID,
				RequestID:     params.RequestID,
			},
			Execution: ExecutionData{
				PartitionKey: params.PartitionKey,
				Priority:     params.Priority,
				Source:       params.Source,
			},
			Schema: SchemaData{
				SchemaVersion: params.SchemaVersion,
				Encoding:      params.Encoding,
			},
			Observability: ObservabilityData{
				Host:      params.Host,
				Timestamp: time.Now().Unix(),
			},
		},
	}
}

// NewNextEvent creates a child event in the parent's aggregate. BFSEngine
// fills the envelope fields inherited from the parent and its version.
func NewNextEvent[T Payload](parent Event, payload T) Event {
	return Event{
		Uuid:          NewUUIDv7(),
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

// NewDerivedEvent creates a child event in a new aggregate of aggType. The
// aggregate ID is derived from the parent event's UUID and purpose, so a
// rebuild reproduces it (ADR-024 rule 2). BFSEngine creates the aggregate and
// fails if it already exists.
func NewDerivedEvent[T Payload](parent Event, aggType enums.AggregateType, purpose string, payload T) Event {
	evt := NewNextEvent(parent, payload)
	evt.AggregateType = aggType
	evt.AggregateUuid = uuidx.NewFromEvent(parent.Uuid, purpose)
	evt.ChildAggregate = ChildAggregateDerived
	return evt
}
