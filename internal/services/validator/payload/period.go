package payload

import (
	"akatengu/internal/model/payload"
	"encoding/json"
	"fmt"
)

func (v *payloadValidator) validateEventPeriodMonthClosed(raw json.RawMessage) error {
	var p payload.PeriodMonthClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ClosedAt == "" {
		errs = append(errs, "closed_at is required")
	}

	return joinErrors(errs)

	return nil
}

func (v *payloadValidator) validateEventPeriodMonthStarted(raw json.RawMessage) error {
	var p payload.PeriodMonthStartedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.PeriodStart == "" {
		errs = append(errs, "period_start is required")
	}

	return joinErrors(errs)
}

func (v *payloadValidator) validateEventPeriodMonthReopened(raw json.RawMessage) error {
	var p payload.PeriodMonthReopenedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ReopenedAt == "" {
		errs = append(errs, "reopened_at is required")
	}

	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}

func (v *payloadValidator) validateEventPeriodAnnualStarted(raw json.RawMessage) error {
	var p payload.PeriodAnnualStartedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.Year == 0 {
		errs = append(errs, "year is required")
	}

	return joinErrors(errs)
}

func (v *payloadValidator) validateEventPeriodAnnualClosed(raw json.RawMessage) error {
	var p payload.PeriodAnnualClosedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ClosedAt == "" {
		errs = append(errs, "closed_at is required")
	}

	return joinErrors(errs)
}

func (v *payloadValidator) validateEventPeriodAnnualReopened(raw json.RawMessage) error {
	var p payload.PeriodAnnualReopenedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ReopenedAt == "" {
		errs = append(errs, "reopened_at is required")
	}

	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}
