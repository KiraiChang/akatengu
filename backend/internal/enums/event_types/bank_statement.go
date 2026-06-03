package event_types

const (
	// bank_csv_template
	EventBankCsvTemplateCreated     eventTypeVal = "bank_csv_template.created"
	EventBankCsvTemplateUpdated     eventTypeVal = "bank_csv_template.updated"
	EventBankCsvTemplateDeactivated eventTypeVal = "bank_csv_template.deactivated"

	// bank_statement
	EventBankStatementImported           eventTypeVal = "bank_statement.imported"
	EventBankStatementTxnMatched         eventTypeVal = "bank_statement.txn_matched"
	EventBankStatementTxnIgnored         eventTypeVal = "bank_statement.txn_ignored"
	EventBankStatementAdjustmentApproved eventTypeVal = "bank_statement.adjustment_approved"
	EventBankStatementCompleted          eventTypeVal = "bank_statement.completed"
)
