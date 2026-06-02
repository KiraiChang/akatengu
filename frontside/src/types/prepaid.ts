export type PrepaidStatus = 'ACTIVE' | 'COMPLETED' | 'DISPOSED';

export interface Prepaid {
  id:                 number;
  prepaid_uuid:       string;
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
  category_uuid: string;
  ledger_uuid:   string;
  name:          string;
  total_amount:  string;
  periods:       number;
  start_date:    string;
  memo:          string;
  note:          string;
}

import type { InstallmentTermsPayload } from './installment';

export interface PrepaidCreatedWithInstallmentPayload {
  category_uuid: string;
  name:          string;
  total_amount:  string;
  periods:       number;
  start_date:    string;
  memo:          string;
  note:          string;
  installment:   InstallmentTermsPayload;
}

export interface PrepaidAmortizedPayload {
  prepaid_uuid: string;
  period_date:  string;
}

export interface PrepaidDisposedPayload {
  prepaid_uuid:  string;
  disposal_date: string;
  memo:          string;
}
