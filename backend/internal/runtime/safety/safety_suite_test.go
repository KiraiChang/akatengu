package safety

import (
	"context"
	"testing"

	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"
	"akatengu/internal/kernel/result"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	typeA event_types.EventType
	typeB event_types.EventType
)

func TestSafetySuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Safety Suite")
}

var _ = BeforeSuite(func() {
	enums.InitEnums()
	typeA = event_types.EventTransactionCreated.Enum()
	typeB = event_types.EventTransactionCorrected.Enum()
})

// ─── Shared test infrastructure ───────────────────────────────────────────────

var ctx = context.Background()

var noop = EventProcessor(func(_ context.Context, _ int, _ event.Event) (result.EventResult, error) {
	return result.EventResult{}, nil
})

func mkEvt(uuid string, typ event_types.EventType, causationID string) event.Event {
	return event.Event{
		Uuid:      uuid,
		EventType: typ,
		Metadata:  event.Metadata{Tracing: event.TracingData{CausationID: causationID}},
	}
}

func withChildren(children []event.Event) EventProcessor {
	return func(_ context.Context, _ int, _ event.Event) (result.EventResult, error) {
		return result.EventResult{Events: children}, nil
	}
}
