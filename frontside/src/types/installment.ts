export type InterestType             = 'FREE' | 'FIXED_RATE';
export type InstallmentStatus        = 'ACTIVE' | 'COMPLETED' | 'CANCELED';
export type InstallmentPaymentStatus = 'PENDING' | 'PAID';

export interface Installment {
  installment_id:    number;
  transaction_id:    number | null;
  ledger_id:         number;
  description:       string;
  total_amount:      string;
  total_periods:     number;
  paid_periods:      number;
  amount_per_period: string;
  start_date:        string;
  end_date:          string | null;
  interest_rate:     string;
  interest_type:     InterestType;
  status:            InstallmentStatus;
  note:              string;
  updated_by:        string | null;
  updated_at:        string | null;
}

export interface InstallmentPayment {
  payment_id:     number;
  installment_id: number;
  transaction_id: number | null;
  period:         number;
  amount:         string;
  interest:       string;
  due_date:       string;
  paid_date:      string | null;
  status:         InstallmentPaymentStatus;
  updated_by:     string | null;
  updated_at:     string | null;
}

export interface InstallmentCreatedPayload {
  amount:            string;
  installment_count: number;
  start_date:        string;
  interest_type:     InterestType;
  annual_rate:       string;
  account_id:        string;
  ledger_id:         number;
  memo:              string;
  note:              string;
}

export interface InstallmentPeriodPaidPayload {
  installment_id: number;
  period:         number;
  paid_date:      string;
  paid_ledger_id: number;
}