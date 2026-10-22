package event_test

import (
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/model/payload"
	"testing"
)

// TestPeriodPayloads_LegacyJSON is the contract of BFS R2 D12 for the period
// opening events; see TestAccountPayloads_LegacyJSON.
func TestPeriodPayloads_LegacyJSON(t *testing.T) {
	t.Run("period.month_started", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.PeriodMonthStartedPayload{PeriodStart: "2026-10"},
			payload.PeriodMonthStartedPayload{PeriodStart: "2026-10"})
	})
	t.Run("period.annual_started", func(t *testing.T) {
		expectSameJSON(t,
			domainevent.PeriodAnnualStartedPayload{Year: 2026},
			payload.PeriodAnnualStartedPayload{Year: 2026})
	})
}
