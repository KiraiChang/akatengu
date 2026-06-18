package engine

import (
	"testing"

	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/kernel/event"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	typeA event_types.EventType
	typeB event_types.EventType
)

func TestBFSEngineSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BFS Engine Suite")
}

var _ = BeforeSuite(func() {
	enums.InitEnums()
	typeA = event_types.EventTransactionCreated.Enum()
	typeB = event_types.EventTransactionCorrected.Enum()
})

func childUUIDs(events []event.Event) []string {
	uuids := make([]string, len(events))
	for i, e := range events {
		uuids[i] = e.Uuid
	}
	return uuids
}
