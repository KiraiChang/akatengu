export interface BankCsvTemplate {
  template_id:        number;
  template_uuid:      string;
  merchant_id:        number;
  template_name:      string;
  encoding:           string;
  skip_rows:          number;
  date_column:        number;
  date_format:        string;
  description_column: number;
  debit_column:       number | null;
  credit_column:      number | null;
  amount_column:      number | null;
  balance_column:     number | null;
  reference_column:   number | null;
  is_active:          boolean;
  note:               string | null;
  version:            number;
  updated_by:         string | null;
  updated_at:         string;
}

export interface CreateBankCsvTemplateRequest {
  template_name:      string;
  encoding:           string;
  skip_rows:          number;
  date_column:        number;
  date_format:        string;
  description_column: number;
  debit_column:       number | null;
  credit_column:      number | null;
  amount_column:      number | null;
  balance_column:     number | null;
  reference_column:   number | null;
  note:               string | null;
}

export interface UpdateBankCsvTemplateRequest extends CreateBankCsvTemplateRequest {
  expected_version: number;
}

export interface BankStatementImport {
  import_id:      number;
  import_uuid:    string;
  merchant_id:    number;
  ledger_id:      number;
  template_id:    number | null;
  statement_date: string;
  import_source:  string;
  filename:       string | null;
  status:         string;
  note:           string | null;
  version:        number;
  updated_by:     string | null;
  updated_at:     string;
}

export interface BankStatementTxn {
  bank_txn_id:      number;
  bank_txn_uuid:    string;
  import_id:        number;
  merchant_id:      number;
  txn_date:         string;
  description:      string;
  debit:            string;
  credit:           string;
  balance:          string | null;
  reference_no:     string | null;
  match_status:     string;
  matched_entry_id: number | null;
  match_confidence: string | null;
  created_txn_id:   number | null;
}

export interface PagedImportsResponse {
  data:  BankStatementImport[];
  total: number;
  page:  number;
  size:  number;
}

export interface ImportResultResponse {
  import:       BankStatementImport;
  transactions: BankStatementTxn[];
}

export interface SuggestedEntry {
  ledger_account_id:  string;
  ledger_id:          number;
  counter_account_id: string;
  amount:             string;
  is_debit:           boolean;
}

export interface ReviewItem {
  bank_txn_id:     number;
  bank_txn_uuid:   string;
  txn_date:        string;
  description:     string;
  debit:           string;
  credit:          string;
  match_status:    string;
  suggested_entry: SuggestedEntry | null;
}

export interface ApproveRequest {
  ledger_id:          number;
  account_id:         string;
  counter_account_id: string;
  txn_date:           string;
  description:        string;
  note:               string | null;
}

// ── PDF Template ─────────────────────────

export interface BankPdfTemplateLedgerItem {
  tpl_ledger_uuid?: string;
  ledger_uuid:      string;
  account_type:     string;
  sort_order:       number;
}

export interface BankPdfTemplateLedgerResponse {
  tpl_ledger_id:   number;
  tpl_ledger_uuid: string;
  template_id:     number;
  ledger_uuid:     string;
  ledger_id:       number | null;
  account_type:    string;
  sort_order:      number;
}

export interface BankPdfTemplate {
  template_id:   number;
  template_uuid: string;
  merchant_id:   number;
  template_name: string;
  bank_type:     string;
  is_active:     boolean;
  version:       number;
  updated_by:    string | null;
  updated_at:    string;
}

export interface CreateBankPdfTemplateRequest {
  template_name: string;
  bank_type:     string;
  ledgers:       BankPdfTemplateLedgerItem[];
}

export interface UpdateBankPdfTemplateRequest extends CreateBankPdfTemplateRequest {
  expected_version: number;
}
