package db

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"encoding/json"
	"time"
)

//go:generate go run ../../../cmd/dbmap-gen

//dbmap:sqlcdb=EventStore
type EventStore struct {
	EventId          int64                 `db:"event_id"`
	EventUuid        string                `db:"event_uuid"`
	OccurredAt       time.Time             `db:"occurred_at"`
	MerchantID       int64                 `db:"merchant_id"`
	AggregateType    enums.AggregateType   `db:"aggregate_type"`
	AggregateId      string                `db:"aggregate_id"`
	AggregateVersion int64                 `db:"aggregate_version"`
	EventType        event_types.EventType `db:"event_type"`
	Payload          json.RawMessage       `db:"payload"`
	Metadata         json.RawMessage       `db:"metadata"`
	UpdatedBy        *string               `db:"updated_by" json:"updated_by"`
}

//dbmap:sqlcdb=AggregateVersion
type AggregateVersion struct {
	AggregateType    enums.AggregateType `db:"aggregate_type"`
	AggregateId      string              `db:"aggregate_id"`
	MerchantID       int64               `db:"merchant_id"`
	AggregateVersion int64               `db:"current_version"`
}

//dbmap:sqlcdb=Snapshot
type Snapshot struct {
	SnapshotId    int64               `db:"snapshot_id"`
	MerchantID    int64               `db:"merchant_id"`
	AggregateType enums.AggregateType `db:"aggregate_type"`
	AggregateId   string              `db:"aggregate_id"`
	AtVersion     int64               `db:"at_version"`
	State         json.RawMessage     `db:"state"`
	CreatedAt     time.Time           `db:"created_at"`
}
