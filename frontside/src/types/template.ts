import type { CashFlowCategory } from './account';

export interface TransactionTemplate {
  id:          number;
  merchant_id: number;
  name:        string;
  description: string | null;
  tag:         string | null;
  updated_by:  string | null;
  updated_at:  string | null;
  version:     number;
}

export interface TransactionTemplateEntry {
  id:                 number;
  merchant_id:        number;
  template_id:        number;
  sort_order:         number;
  account_id:         string;
  ledger_id:          number | null;
  debit:              string;
  credit:             string;
  note:               string | null;
  cash_flow_category: CashFlowCategory | null;
}

export interface TransactionTemplateDetail extends TransactionTemplate {
  entries: TransactionTemplateEntry[];
}

export interface SaveTemplateEntryRequest {
  sort_order:         number;
  account_id:         string;
  ledger_id:          number | null;
  debit:              number;
  credit:             number;
  note:               string | null;
  cash_flow_category: CashFlowCategory | null;
}

export interface SaveTemplateRequest {
  name:        string;
  description: string | null;
  tag:         string | null;
  entries:     SaveTemplateEntryRequest[];
}
