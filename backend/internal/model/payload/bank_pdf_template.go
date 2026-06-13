package payload

import "fmt"

// TemplateLedgerItem 用於模板事件中定義帳本對應清單。
type TemplateLedgerItem struct {
	TplLedgerUUID string `json:"tpl_ledger_uuid"`
	LedgerUUID    string `json:"ledger_uuid"`
	AccountType   string `json:"account_type"`
	SortOrder     int    `json:"sort_order"`
}

// ─────────────────────────────────────────
// BankPdfTemplate payloads
// ─────────────────────────────────────────

type BankPdfTemplateCreatedPayload struct {
	TemplateUUID string               `json:"template_uuid"`
	TemplateName string               `json:"template_name"`
	BankType     string               `json:"bank_type"`
	Ledgers      []TemplateLedgerItem `json:"ledgers"`
}

func (p BankPdfTemplateCreatedPayload) Validate() error {
	var errs []string
	if p.TemplateUUID == "" {
		errs = append(errs, "template_uuid is required")
	}
	if p.TemplateName == "" {
		errs = append(errs, "template_name is required")
	}
	if p.BankType == "" {
		errs = append(errs, "bank_type is required")
	}
	if len(p.Ledgers) == 0 {
		errs = append(errs, "ledgers cannot be empty")
	}
	for i, l := range p.Ledgers {
		if l.TplLedgerUUID == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].tpl_ledger_uuid is required", i))
		}
		if l.LedgerUUID == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].ledger_uuid is required", i))
		}
		if l.AccountType == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].account_type is required", i))
		}
	}
	return joinErrors(errs)
}

type BankPdfTemplateUpdatedPayload struct {
	TemplateUUID string               `json:"template_uuid"`
	TemplateName string               `json:"template_name"`
	BankType     string               `json:"bank_type"`
	Ledgers      []TemplateLedgerItem `json:"ledgers"`
}

func (p BankPdfTemplateUpdatedPayload) Validate() error {
	var errs []string
	if p.TemplateUUID == "" {
		errs = append(errs, "template_uuid is required")
	}
	if p.TemplateName == "" {
		errs = append(errs, "template_name is required")
	}
	if p.BankType == "" {
		errs = append(errs, "bank_type is required")
	}
	if len(p.Ledgers) == 0 {
		errs = append(errs, "ledgers cannot be empty")
	}
	for i, l := range p.Ledgers {
		if l.TplLedgerUUID == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].tpl_ledger_uuid is required", i))
		}
		if l.LedgerUUID == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].ledger_uuid is required", i))
		}
		if l.AccountType == "" {
			errs = append(errs, fmt.Sprintf("ledgers[%d].account_type is required", i))
		}
	}
	return joinErrors(errs)
}

type BankPdfTemplateDeactivatedPayload struct {
	TemplateUUID string `json:"template_uuid"`
}

func (p BankPdfTemplateDeactivatedPayload) Validate() error {
	var errs []string
	if p.TemplateUUID == "" {
		errs = append(errs, "template_uuid is required")
	}
	return joinErrors(errs)
}
