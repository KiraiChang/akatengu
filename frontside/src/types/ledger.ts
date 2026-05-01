export type LedgerAccountType = 'BANK_ACCOUNT' | 'CREDIT_CARD' | 'LOAN';

export interface LedgerAccount {
  ledger_id:    number;
  account_id:   string;
  institution:  string;
  name:         string;
  type:         LedgerAccountType;
  account_no:   string | null;
  currency:     string;
  credit_limit: string | null;
  billing_day:  string | null;
  due_day:      string | null;
  is_active:    boolean;
  note:         string | null;
  version:      number;
}

export interface LedgerAccountUpdatePayload {
  ledger_id:    number;
  account_id:   string;
  institution:  string;
  name:         string;
  type:         LedgerAccountType;
  account_no:   string | null;
  currency:     string;
  credit_limit: string | null;
  billing_day:  string | null;
  due_day:      string | null;
  is_active:    boolean;
  note:         string | null;
  version:      number;
}

export interface LedgerBalance {
  ledger_id: number;
  balance:   string;
}

export interface LedgerAccountCreatePayload {
  account_id:   string;
  institution:  string;
  name:         string;
  type:         LedgerAccountType;
  account_no:   string | null;
  currency:     string;
  credit_limit: string | null;
  billing_day:  string | null;
  due_day:      string | null;
  is_active:    boolean;
  note:         string | null;
}
