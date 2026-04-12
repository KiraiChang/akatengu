package event_types

const (
	// installment
	EventInstallmentCreated    eventTypeVal = "installment.created"
	EventInstallmentPeriodPaid eventTypeVal = "installment.period_paid"
	EventInstallmentCompleted  eventTypeVal = "installment.completed"
	EventInstallmentCancelled  eventTypeVal = "installment.cancelled"
)
