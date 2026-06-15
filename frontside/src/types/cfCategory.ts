import type { CashFlowCategory } from './account';

export interface CFReviewRow {
  txn_id:       number;
  txn_uuid:     string;
  txn_date:     string;
  description:  string;
  total_amount: string;
  currency:     string;
  entry_id:     number;
  entry_uuid:   string;
  account_id:   string;
  ledger_id:    number | null;
  debit:        string;
  credit:       string;
  cf_category:  CashFlowCategory | null;
  is_confirmed: boolean;
}

export interface CFReviewTxn {
  txn_id:       number;
  txn_uuid:     string;
  txn_date:     string;
  description:  string;
  total_amount: string;
  currency:     string;
  entries:      CFReviewRow[];
}

export type CFActivityCategory = Exclude<CashFlowCategory, 'CASH'>;
