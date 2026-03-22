package enums

import "akatengu/internal/pkg/enumx"

type EventType = enumx.EnumVal[eventTypeVal]
type eventTypeVal string

const (
	// transaction
	eventTransactionCreated   eventTypeVal = "transaction.created"
	eventTransactionCorrected eventTypeVal = "transaction.corrected"
	eventTransactionVoided    eventTypeVal = "transaction.voided"

	// installment
	eventInstallmentCreated    eventTypeVal = "installment.created"
	eventInstallmentPeriodPaid eventTypeVal = "installment.period_paid"
	eventInstallmentCompleted  eventTypeVal = "installment.completed"
	eventInstallmentCancelled  eventTypeVal = "installment.cancelled"

	// reconciliation
	eventReconciliationStarted         eventTypeVal = "reconciliation.started"
	eventReconciliationAdjustmentAdded eventTypeVal = "reconciliation.adjustment_added"
	eventReconciliationBalanced        eventTypeVal = "reconciliation.balanced"
	eventReconciliationReopened        eventTypeVal = "reconciliation.reopened"

	// account
	eventAccountCreated           eventTypeVal = "account.created"
	eventAccountUpdated           eventTypeVal = "account.updated"
	eventAccountDeactivated       eventTypeVal = "account.deactivated"
	eventLedgerAccountCreated     eventTypeVal = "ledger_account.created"
	eventLedgerAccountUpdated     eventTypeVal = "ledger_account.updated"
	eventLedgerAccountDeactivated eventTypeVal = "ledger_account.deactivated"
)

var validEventTypes = []eventTypeVal{
	eventTransactionCreated, eventTransactionCorrected, eventTransactionVoided,
	eventInstallmentCreated, eventInstallmentPeriodPaid, eventInstallmentCompleted, eventInstallmentCancelled,
	eventReconciliationStarted, eventReconciliationAdjustmentAdded, eventReconciliationBalanced, eventReconciliationReopened,
	eventAccountCreated, eventAccountUpdated, eventAccountDeactivated, eventLedgerAccountCreated, eventLedgerAccountUpdated, eventLedgerAccountDeactivated,
}

var (
	// transaction
	EventTransactionCreated   = enumx.Must(string(eventTransactionCreated), validEventTypes)
	EventTransactionCorrected = enumx.Must(string(eventTransactionCorrected), validEventTypes)
	EventTransactionVoided    = enumx.Must(string(eventTransactionVoided), validEventTypes)

	// installment
	EventInstallmentCreated    = enumx.Must(string(eventInstallmentCreated), validEventTypes)
	EventInstallmentPeriodPaid = enumx.Must(string(eventInstallmentPeriodPaid), validEventTypes)
	EventInstallmentCompleted  = enumx.Must(string(eventInstallmentCompleted), validEventTypes)
	EventInstallmentCancelled  = enumx.Must(string(eventInstallmentCancelled), validEventTypes)

	// reconciliation
	EventReconciliationStarted         = enumx.Must(string(eventReconciliationStarted), validEventTypes)
	EventReconciliationAdjustmentAdded = enumx.Must(string(eventReconciliationAdjustmentAdded), validEventTypes)
	EventReconciliationBalanced        = enumx.Must(string(eventReconciliationBalanced), validEventTypes)
	EventReconciliationReopened        = enumx.Must(string(eventReconciliationReopened), validEventTypes)

	// account
	EventAccountCreated           = enumx.Must(string(eventAccountCreated), validEventTypes)
	EventAccountUpdated           = enumx.Must(string(eventAccountUpdated), validEventTypes)
	EventAccountDeactivated       = enumx.Must(string(eventAccountDeactivated), validEventTypes)
	EventLedgerAccountCreated     = enumx.Must(string(eventLedgerAccountCreated), validEventTypes)
	EventLedgerAccountUpdated     = enumx.Must(string(eventLedgerAccountUpdated), validEventTypes)
	EventLedgerAccountDeactivated = enumx.Must(string(eventLedgerAccountDeactivated), validEventTypes)
)

func ParseEventType(s string) (EventType, error) {
	return enumx.New(s, validEventTypes)
}
