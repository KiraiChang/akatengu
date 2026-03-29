package payload

import (
	"akatengu/internal/model/payload"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

func validateTransactionCreated(raw json.RawMessage) error {
	var p payload.TransactionCreatedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string

	if p.TransactionDate == "" {
		errs = append(errs, "txn_date is required")
	} else if _, err := time.Parse("2006-01-02", p.TransactionDate); err != nil {
		errs = append(errs, "txn_date must be YYYY-MM-DD")
	}
	if p.Description == "" {
		errs = append(errs, "description is required")
	}
	if !p.TotalAmount.IsPositive() {
		errs = append(errs, "total_amount must be > 0")
	}
	if len(p.Entries) < 2 {
		errs = append(errs, "entries must have at least 2 items")
	}

	// 借貸必須平衡
	var totalDebit, totalCredit decimal.Decimal
	for i, e := range p.Entries {
		if e.AccountId == "" {
			errs = append(errs, fmt.Sprintf("entries[%d].account_id is required", i))
		}
		totalDebit = totalDebit.Add(e.Debit)
		totalCredit = totalCredit.Add(e.Credit)
	}
	if !totalDebit.Sub(totalCredit).IsZero() {
		errs = append(errs, fmt.Sprintf(
			"entries not balanced: debit=%.2f credit=%.2f", totalDebit, totalCredit,
		))
	}

	return joinErrors(errs)
}

func validateTransactionCorrected(raw json.RawMessage) error {
	var p payload.TransactionCorrectedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string
	if p.OriginalTransactionId <= 0 {
		errs = append(errs, "original_txn_id is required")
	}
	if p.VoidTransactionId <= 0 {
		errs = append(errs, "void_txn_id is required")
	}
	if p.CorrectedTransactionId <= 0 {
		errs = append(errs, "correction_txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}
	return joinErrors(errs)
}

func validateTransactionVoided(raw json.RawMessage) error {
	var p payload.TransactionVoidedPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return fmt.Errorf("malformed payloadValidator: %w", err)
	}

	var errs []string
	if p.TransactionId <= 0 {
		errs = append(errs, "txn_id is required")
	}
	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}
	return joinErrors(errs)
}
