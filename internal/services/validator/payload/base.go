package payload

import (
	"akatengu/internal/model/enums/event_types"
	"akatengu/internal/services/validator"
	"context"

	"encoding/json"
	"fmt"
	"strings"
)

type eventValidator struct {
	rules map[string]validateFn
}

type validateFn func(payload json.RawMessage) error

func NewValidator() validator.Validator {
	v := &eventValidator{
		rules: map[string]validateFn{},
	}
	v.register()
	return v
}

func (v *eventValidator) Validate(ctx context.Context, eventType event_types.EventType, payload json.RawMessage) error {
	fn, ok := v.rules[eventType.String()]
	if !ok {
		return fmt.Errorf("unknown event type: %s", eventType)
	}
	return fn(payload)
}

func (v *eventValidator) register() {
	// transaction
	v.rules[event_types.EventTransactionCreated.String()] = validateTransactionCreated
	v.rules[event_types.EventTransactionCorrected.String()] = validateTransactionCorrected
	v.rules[event_types.EventTransactionVoided.String()] = validateTransactionVoided

	//// installment
	//v.rules[enums.EventInstallmentCreated.String()] = validateInstallmentCreated
	//v.rules[enums.EventInstallmentPeriodPaid.String()] = validateInstallmentPeriodPaid
	//
	//// reconciliation
	//v.rules[enums.EventReconciliationStarted.String()] = validateReconciliationStarted
	//v.rules[enums.EventReconciliationAdjustmentAdded.String()] = validateReconciliationAdjustmentAdded

	// account
	v.rules[event_types.EventAccountCreated.String()] = validateAccountCreate
	v.rules[event_types.EventAccountUpdated.String()] = validateAccountUpdated
	v.rules[event_types.EventLedgerAccountCreated.String()] = validateLedgerAccountCreate
	v.rules[event_types.EventLedgerAccountUpdated.String()] = validateLedgerAccountUpdated

	// investment
	v.rules[event_types.EventInvestmentCreate.String()] = validateInvestmentCreate
	v.rules[event_types.EventInvestmentUpdate.String()] = validateInvestmentUpdate
}

func joinErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("validation failed: %s", strings.Join(errs, "; "))
}
