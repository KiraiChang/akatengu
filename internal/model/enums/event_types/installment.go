package event_types

import "akatengu/internal/pkg/enumx"

const (
	// installment
	eventInstallmentCreated    eventTypeVal = "installment.created"
	eventInstallmentPeriodPaid eventTypeVal = "installment.period_paid"
	eventInstallmentCompleted  eventTypeVal = "installment.completed"
	eventInstallmentCancelled  eventTypeVal = "installment.cancelled"
)

var validInstallmentEventTypes = []eventTypeVal{
	eventInstallmentCreated,
	eventInstallmentPeriodPaid,
	eventInstallmentCompleted,
	eventInstallmentCancelled,
}

var (
	// installment
	EventInstallmentCreated    = enumx.Must(string(eventInstallmentCreated), validEventTypes)
	EventInstallmentPeriodPaid = enumx.Must(string(eventInstallmentPeriodPaid), validEventTypes)
	EventInstallmentCompleted  = enumx.Must(string(eventInstallmentCompleted), validEventTypes)
	EventInstallmentCancelled  = enumx.Must(string(eventInstallmentCancelled), validEventTypes)
)
