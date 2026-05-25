package request

import (
	"akatengu/internal/enums"
	"akatengu/internal/repos/query"
	"errors"

	"github.com/shopspring/decimal"
)

type TemplateEntryRequest struct {
	SortOrder        int64                   `json:"sort_order"`
	AccountID        string                  `json:"account_id"`
	LedgerID         *int64                  `json:"ledger_id"`
	Debit            decimal.Decimal         `json:"debit"`
	Credit           decimal.Decimal         `json:"credit"`
	Note             *string                 `json:"note"`
	CashFlowCategory *enums.CashFlowCategory `json:"cash_flow_category"`
}

type SaveTemplateRequest struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description"`
	Tag         *string                `json:"tag"`
	Entries     []TemplateEntryRequest `json:"entries"`
}

func (r *SaveTemplateRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	for i, e := range r.Entries {
		if e.AccountID == "" {
			return errors.New("entry account_id is required")
		}
		_ = i
	}
	return nil
}

func (r *SaveTemplateRequest) ToCreateCmd(updatedBy *string) query.CreateTemplateCmd {
	return query.CreateTemplateCmd{
		Name:        r.Name,
		Description: r.Description,
		Tag:         r.Tag,
		UpdatedBy:   updatedBy,
		Entries:     r.toEntryInputs(),
	}
}

func (r *SaveTemplateRequest) ToUpdateCmd(updatedBy *string) query.UpdateTemplateCmd {
	return query.UpdateTemplateCmd{
		Name:        r.Name,
		Description: r.Description,
		Tag:         r.Tag,
		UpdatedBy:   updatedBy,
		Entries:     r.toEntryInputs(),
	}
}

func (r *SaveTemplateRequest) toEntryInputs() []query.TemplateEntryInput {
	inputs := make([]query.TemplateEntryInput, len(r.Entries))
	for i, e := range r.Entries {
		var cf enums.CashFlowCategory
		if e.CashFlowCategory != nil {
			cf = *e.CashFlowCategory
		}
		inputs[i] = query.TemplateEntryInput{
			SortOrder:        e.SortOrder,
			AccountID:        e.AccountID,
			LedgerID:         e.LedgerID,
			Debit:            e.Debit,
			Credit:           e.Credit,
			Note:             e.Note,
			CashFlowCategory: cf,
		}
	}
	return inputs
}
