package event_types

const (
	// reconciliation
	EventReconciliationStarted         eventTypeVal = "reconciliation.started"
	EventReconciliationAdjustmentAdded eventTypeVal = "reconciliation.adjustment_added"
	EventReconciliationBalanced        eventTypeVal = "reconciliation.balanced"
	EventReconciliationReopened        eventTypeVal = "reconciliation.reopened"
)
