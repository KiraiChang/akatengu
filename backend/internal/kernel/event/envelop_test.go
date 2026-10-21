package event_test

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/pkg/uuidx"
	"akatengu/internal/shared/utils/test"
	"testing"
)

type derivedTestPayload struct{}

func (derivedTestPayload) EventType() event_types.EventType {
	return event_types.EventTransactionCorrected.Enum()
}

func TestNewDerivedEvent(t *testing.T) {
	test.InitEnums()

	const (
		parentA = "0192f3e0-0000-7000-8000-000000000001"
		parentB = "0192f3e0-0000-7000-8000-000000000002"
	)
	parent := func(uuid string) event.Event {
		p := event.NewEvent[derivedTestPayload]("parent-aggregate", derivedTestPayload{}, event.Metadata{
			Tracing: event.TracingData{CorrelationID: "0192f3e0-0000-7000-8000-0000000000c1"},
		})
		p.Uuid = uuid
		p.AggregateType = enums.AggregateAccount.Enum()
		return p
	}
	aggType := enums.AggregateTransaction.Enum()

	cases := []struct {
		name              string
		parentA, purposeA string
		parentB, purposeB string
		wantSameAggregate bool
	}{
		{"相同的父事件與用途得到相同的 aggregate ID", parentA, "txn:closing", parentA, "txn:closing", true},
		{"不同用途得到不同的 aggregate ID", parentA, "txn:closing", parentA, "txn:opening", false},
		{"不同父事件得到不同的 aggregate ID", parentA, "txn:closing", parentB, "txn:closing", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pa := parent(tc.parentA)
			a := event.NewDerivedEvent(pa, aggType, tc.purposeA, derivedTestPayload{})
			b := event.NewDerivedEvent(parent(tc.parentB), aggType, tc.purposeB, derivedTestPayload{})

			if want := uuidx.NewFromEvent(tc.parentA, tc.purposeA); a.AggregateUuid != want {
				t.Fatalf("aggregate ID = %s, want %s", a.AggregateUuid, want)
			}
			if same := a.AggregateUuid == b.AggregateUuid; same != tc.wantSameAggregate {
				t.Fatalf("same aggregate = %v, want %v (%s, %s)", same, tc.wantSameAggregate, a.AggregateUuid, b.AggregateUuid)
			}
			if a.AggregateType != aggType {
				t.Errorf("aggregate type = %v, want %v", a.AggregateType, aggType)
			}
			if a.ChildAggregate != event.ChildAggregateDerived {
				t.Errorf("ChildAggregate = %v, want ChildAggregateDerived", a.ChildAggregate)
			}
			if a.Metadata.Tracing.CausationID != pa.Uuid {
				t.Errorf("causation ID = %s, want the parent UUID %s", a.Metadata.Tracing.CausationID, pa.Uuid)
			}
			if a.Uuid == "" || a.Uuid == pa.Uuid {
				t.Errorf("event UUID = %q, want a new UUID", a.Uuid)
			}
		})
	}
}
