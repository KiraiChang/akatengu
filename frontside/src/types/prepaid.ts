export type PrepaidStatus = 'ACTIVE' | 'COMPLETED' | 'DISPOSED';

export interface Prepaid {
  id:                 number;
  merchant_id:        number;
  txn_id:             number | null;
  account_id:         string;
  expense_account_id: string;
  name:               string;
  total_amount:       string;
  amortized_amount:   string;
  periods:            number;
  amortized_periods:  number;
  start_date:         string;
  status:             PrepaidStatus;
  updated_by:         string | null;
  updated_at:         string | null;
  version:            number;
}

export interface PrepaidAmortization {
  id:          number;
  merchant_id: number;
  prepaid_id:  number;
  txn_id:      number;
  period_date: string;
  amount:      string;
  updated_by:  string | null;
  updated_at:  string | null;
}

export interface PrepaidCreatedPayload {
  account_id:         string;
  expense_account_id: string;
  ledger_id:          number;
  name:               string;
  total_amount:       string;
  periods:            number;
  start_date:         string;
  memo:               string;
  note:               string;
}

export interface PrepaidAmortizedPayload {
  prepaid_id:  number;
  period_date: string;
}

export interface PrepaidDisposedPayload {
  prepaid_id:    number;
  disposal_date: string;
  memo:          string;
}
