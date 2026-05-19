import type { CashFlowCategory } from './account';

export type TransactionStatus = 'ACTIVE' | 'CORRECTED' | 'VOIDED' | 'VOID_REF';

export interface Transaction {
  txn_id: number;
  txn_date: string;
  description: string;
  total_amount: string;
  currency: string;
  status: TransactionStatus;
  installment_id: number | null;
  receipt_no: string | null;
  note: string | null;
  version: number;
  ref_txn_id: number | null;
  updated_by: string | null;
  updated_at: string | null;
}

export interface Entry {
  entry_id:           number;
  txn_id:             number;
  ledger_id:          number | null;
  account_id:         string;
  debit:              string;
  credit:             string;
  note:               string | null;
  cash_flow_category: CashFlowCategory | null;
  updated_by:         string | null;
  updated_at:         string | null;
}

export interface TransactionEntryPayload {
  account_id:         string;
  ledger_id:          number | null;
  debit:              number;
  credit:             number;
  cash_flow_category: CashFlowCategory | null;
}

export interface TransactionCreatedPayload {
  transaction_date: string;
  description: string;
  total_amount: number;
  currency: string;
  receipt_no: string | null;
  note: string | null;
  ref_txn_id: number | null;
  entries: TransactionEntryPayload[];
}
