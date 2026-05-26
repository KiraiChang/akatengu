package projection

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"encoding/json"
)

// All audit types use manual mapping in the repo layer (no dbmap annotation).
// ExchangeRate is already defined in investment.go — reused here.

type AggregateVersionAudit struct {
	AggregateType  enums.AggregateType `json:"aggregate_type"`
	AggregateId    string              `json:"aggregate_id"`
	MerchantID     int64               `json:"merchant_id"`
	CurrentVersion int64               `json:"current_version"`
}

type EventStoreAudit struct {
	EventId          int64                 `json:"event_id"`
	EventUuid        string                `json:"event_uuid"`
	OccurredAt       string                `json:"occurred_at"`
	MerchantID       int64                 `json:"merchant_id"`
	AggregateType    enums.AggregateType   `json:"aggregate_type"`
	AggregateId      string                `json:"aggregate_id"`
	AggregateVersion int64                 `json:"aggregate_version"`
	EventType        event_types.EventType `json:"event_type"`
	Payload          json.RawMessage       `json:"payload"`
	Metadata         *json.RawMessage      `json:"metadata"`
	UpdatedBy        *string               `json:"updated_by"`
}

type CheckpointAudit struct {
	ProjectionName string  `json:"projection_name"`
	MerchantID     int64   `json:"merchant_id"`
	LastEventID    int64   `json:"last_event_id"`
	UpdatedAt      *string `json:"updated_at"`
}

type SnapshotAudit struct {
	SnapshotId    int64               `json:"snapshot_id"`
	MerchantID    int64               `json:"merchant_id"`
	AggregateType enums.AggregateType `json:"aggregate_type"`
	AggregateId   string              `json:"aggregate_id"`
	AtVersion     int64               `json:"at_version"`
	State         json.RawMessage     `json:"state"`
	CreatedAt     string              `json:"created_at"`
}
