package event

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
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
