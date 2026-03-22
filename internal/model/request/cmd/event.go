package cmd

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/model/enums/event_types"
)

// AppendCmd 是寫入事件時傳入的指令
type AppendCmd struct {
	AggregateType   enums.AggregateType   `json:"aggregate_type"`
	AggregateID     string                `json:"aggregate_id"`
	ExpectedVersion int64                 `json:"expected_version"`
	EventType       event_types.EventType `json:"event_type"`
	Payload         interface{}           `json:"payload"`
	Metadata        interface{}           `json:"metadata"`
}
