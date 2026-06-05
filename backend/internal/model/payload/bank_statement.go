package payload

import "github.com/shopspring/decimal"

// ─────────────────────────────────────────
// BankCsvTemplate payloads
// ─────────────────────────────────────────

type BankCsvTemplateCreatedPayload struct {
	TemplateName      string  `json:"template_name"`
	Encoding          string  `json:"encoding"`
	SkipRows          int64   `json:"skip_rows"`
	DateColumn        int64   `json:"date_column"`
	DateFormat        string  `json:"date_format"`
	DescriptionColumn int64   `json:"description_column"`
	DebitColumn       *int64  `json:"debit_column"`
	CreditColumn      *int64  `json:"credit_column"`
	AmountColumn      *int64  `json:"amount_column"`
	BalanceColumn     *int64  `json:"balance_column"`
	ReferenceColumn   *int64  `json:"reference_column"`
	Note              *string `json:"note"`
}

func (p BankCsvTemplateCreatedPayload) Validate() error {
	var errs []string
	if p.TemplateName == "" {
		errs = append(errs, "template_name is required")
	}
	if p.DateFormat == "" {
		errs = append(errs, "date_format is required")
	}
	if p.DebitColumn == nil && p.CreditColumn == nil && p.AmountColumn == nil {
		errs = append(errs, "at least one of debit_column, credit_column, or amount_column is required")
	}
	return joinErrors(errs)
}

type BankCsvTemplateUpdatedPayload struct {
	TemplateUUID      string  `json:"template_uuid"`
	TemplateName      string  `json:"template_name"`
	Encoding          string  `json:"encoding"`
	SkipRows          int64   `json:"skip_rows"`
	DateColumn        int64   `json:"date_column"`
	DateFormat        string  `json:"date_format"`
	DescriptionColumn int64   `json:"description_column"`
	DebitColumn       *int64  `json:"debit_column"`
	CreditColumn      *int64  `json:"credit_column"`
	AmountColumn      *int64  `json:"amount_column"`
	BalanceColumn     *int64  `json:"balance_column"`
	ReferenceColumn   *int64  `json:"reference_column"`
	Note              *string `json:"note"`
}

func (p BankCsvTemplateUpdatedPayload) Validate() error {
	var errs []string
	if p.TemplateUUID == "" {
		errs = append(errs, "template_uuid is required")
	}
	if p.TemplateName == "" {
		errs = append(errs, "template_name is required")
	}
	if p.DateFormat == "" {
		errs = append(errs, "date_format is required")
	}
	if p.DebitColumn == nil && p.CreditColumn == nil && p.AmountColumn == nil {
		errs = append(errs, "at least one of debit_column, credit_column, or amount_column is required")
	}
	return joinErrors(errs)
}

type BankCsvTemplateDeactivatedPayload struct {
	TemplateUUID string `json:"template_uuid"`
}

func (p BankCsvTemplateDeactivatedPayload) Validate() error {
	var errs []string
	if p.TemplateUUID == "" {
		errs = append(errs, "template_uuid is required")
	}
	return joinErrors(errs)
}

// ─────────────────────────────────────────
// BankStatement payloads
// ─────────────────────────────────────────

// BankStatementTxnItem is a single bank transaction row from the CSV.
type BankStatementTxnItem struct {
	BankTxnUUID string          `json:"bank_txn_uuid"`
	TxnDate     string          `json:"txn_date"`
	Description string          `json:"description"`
	Debit       decimal.Decimal `json:"debit"`
	Credit      decimal.Decimal `json:"credit"`
	Balance     *decimal.Decimal `json:"balance"`
	ReferenceNo *string         `json:"reference_no"`
}

type BankStatementImportedPayload struct {
	LedgerID      int64                  `json:"ledger_id"`
	TemplateID    *int64                 `json:"template_id"`
	StatementDate string                 `json:"statement_date"`
	ImportSource  string                 `json:"import_source"`
	Filename      *string                `json:"filename"`
	Note          *string                `json:"note"`
	Transactions  []BankStatementTxnItem `json:"transactions"`
}

func (p BankStatementImportedPayload) Validate() error {
	var errs []string
	if p.LedgerID == 0 {
		errs = append(errs, "ledger_id is required")
	}
	if p.StatementDate == "" {
		errs = append(errs, "statement_date is required")
	}
	if len(p.Transactions) == 0 {
		errs = append(errs, "transactions cannot be empty")
	}
	return joinErrors(errs)
}

// BankStatementTxnMatchedPayload is emitted when auto/manual matching updates a bank txn.
type BankStatementTxnMatchedPayload struct {
	BankTxnID       int64  `json:"bank_txn_id"`
	MatchedEntryID  *int64 `json:"matched_entry_id"`
	MatchStatus     string `json:"match_status"`
	MatchConfidence string `json:"match_confidence"`
}

func (p BankStatementTxnMatchedPayload) Validate() error {
	var errs []string
	if p.BankTxnID == 0 {
		errs = append(errs, "bank_txn_id is required")
	}
	if p.MatchStatus == "" {
		errs = append(errs, "match_status is required")
	}
	return joinErrors(errs)
}

// BankStatementTxnIgnoredPayload is emitted when a user ignores an unmatched bank txn.
type BankStatementTxnIgnoredPayload struct {
	BankTxnID int64 `json:"bank_txn_id"`
}

func (p BankStatementTxnIgnoredPayload) Validate() error {
	var errs []string
	if p.BankTxnID == 0 {
		errs = append(errs, "bank_txn_id is required")
	}
	return joinErrors(errs)
}

// BankStatementAdjustmentApprovedPayload is emitted when user approves a new transaction from unmatched bank txn.
type BankStatementAdjustmentApprovedPayload struct {
	BankTxnID    int64   `json:"bank_txn_id"`
	CreatedTxnID int64   `json:"created_txn_id"`
	Description  string  `json:"description"`
	Note         *string `json:"note"`
}

func (p BankStatementAdjustmentApprovedPayload) Validate() error {
	var errs []string
	if p.BankTxnID == 0 {
		errs = append(errs, "bank_txn_id is required")
	}
	if p.CreatedTxnID == 0 {
		errs = append(errs, "created_txn_id is required")
	}
	return joinErrors(errs)
}

// BankStatementCompletedPayload is emitted when all items in an import are reviewed.
type BankStatementCompletedPayload struct {
	ImportUUID string `json:"import_uuid"`
}

func (p BankStatementCompletedPayload) Validate() error {
	var errs []string
	if p.ImportUUID == "" {
		errs = append(errs, "import_uuid is required")
	}
	return joinErrors(errs)
}
