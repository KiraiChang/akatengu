package event_types

import "akatengu/internal/pkg/enumx"

const (
	// reconciliation
	eventReconciliationStarted         eventTypeVal = "reconciliation.started"
	eventReconciliationAdjustmentAdded eventTypeVal = "reconciliation.adjustment_added"
	eventReconciliationBalanced        eventTypeVal = "reconciliation.balanced"
	eventReconciliationReopened        eventTypeVal = "reconciliation.reopened"
)

var validReconciliationEventTypes = []eventTypeVal{
	eventReconciliationStarted,
	eventReconciliationAdjustmentAdded,
	eventReconciliationBalanced,
	eventReconciliationReopened,
}

var (
	// reconciliation
	EventReconciliationStarted         = enumx.Must(string(eventReconciliationStarted), validEventTypes)
	EventReconciliationAdjustmentAdded = enumx.Must(string(eventReconciliationAdjustmentAdded), validEventTypes)
	EventReconciliationBalanced        = enumx.Must(string(eventReconciliationBalanced), validEventTypes)
	EventReconciliationReopened        = enumx.Must(string(eventReconciliationReopened), validEventTypes)
)
