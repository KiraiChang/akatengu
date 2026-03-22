package event_types

import "akatengu/internal/pkg/enumx"

const (
	// transaction
	eventTransactionCreated   eventTypeVal = "transaction.created"
	eventTransactionCorrected eventTypeVal = "transaction.corrected"
	eventTransactionVoided    eventTypeVal = "transaction.voided"
)

var validTransactionEventTypes = []eventTypeVal{
	eventTransactionCreated,
	eventTransactionCorrected,
	eventTransactionVoided,
}

var (
	// transaction
	EventTransactionCreated   = enumx.Must(string(eventTransactionCreated), validEventTypes)
	EventTransactionCorrected = enumx.Must(string(eventTransactionCorrected), validEventTypes)
	EventTransactionVoided    = enumx.Must(string(eventTransactionVoided), validEventTypes)
)
