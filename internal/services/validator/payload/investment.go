package payload

import (
	"akatengu/internal/model/payload"
	"encoding/json"
	"fmt"
)

func validateInvestmentUpdate(raw json.RawMessage) error {
	var p payload.InvestmentUpdatedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.InvestmentId == 0 {
		errs = append(errs, "investment_id is required")
	}

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
}

func validateInvestmentCreate(raw json.RawMessage) error {
	var p payload.InvestmentCreatedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.AccountId == "" {
		errs = append(errs, "account_id is required")
	}
	if p.Symbol == "" {
		errs = append(errs, "symbol is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
	}

	return joinErrors(errs)
}
