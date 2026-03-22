package payload

import (
	"akatengu/internal/model/payload"
	"encoding/json"
	"fmt"
)

func validateAccountCreate(raw json.RawMessage) error {
	var p payload.AccountCreatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed eventValidator: %w", err)
	}

	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "account name is required")
	}

	return joinErrors(errs)
}

func validateAccountUpdated(raw json.RawMessage) error {
	var p payload.AccountUpdatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed eventValidator: %w", err)
	}

	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "account name is required")
	}

	if p.Version == 0 {
		errs = append(errs, "version is required")
	}

	return joinErrors(errs)
}

func validateLedgerAccountCreate(raw json.RawMessage) error {
	var p payload.LedgerAccountCreatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed eventValidator: %w", err)
	}

	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "ledger name is required")
	}

	if p.Currency == "" {
		errs = append(errs, "ledger currency is required")
	}

	if p.Institution == "" {
		errs = append(errs, "ledger institution is required")
	}

	return joinErrors(errs)
}

func validateLedgerAccountUpdated(raw json.RawMessage) error {
	var p payload.LedgerAccountUpdatePayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed eventValidator: %w", err)
	}

	var errs []string

	if p.LedgerId == 0 {
		errs = append(errs, "ledger id is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account id is required")
	}

	if p.Name == "" {
		errs = append(errs, "ledger name is required")
	}

	if p.Currency == "" {
		errs = append(errs, "ledger currency is required")
	}

	if p.Institution == "" {
		errs = append(errs, "ledger institution is required")
	}

	if p.Version == 0 {
		errs = append(errs, "ledger version is required")
	}

	return joinErrors(errs)
}
