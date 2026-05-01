export type PeriodType   = 'MONTHLY' | 'ANNUAL';
export type PeriodStatus = 'OPEN' | 'CLOSED' | 'REOPENED';

// GET response — PeriodClosing 無 json tags → PascalCase
export interface PeriodClosing {
  closing_id:     number;
  period_type:    PeriodType;
  period_start:   string;
  period_end:     string;
  status:         PeriodStatus;
  opening_txn_id: number | null;
  closing_txn_id: number | null;
  snapshot:       string | null;
  closed_at:      string | null;
  note:           string | null;
  reopen_at:      string | null;
  reopen_reason:  string | null;
}

// POST payloads — 有 json tags → snake_case
export interface PeriodMonthStartedPayload {
  period_start: string;
}

export interface PeriodMonthClosedPayload {
  closing_id: number;
  closed_at:  string;
}

export interface PeriodMonthReopenedPayload {
  closing_id:   number;
  reason:       string;
  reopened_at:  string;
}

export interface PeriodAnnualStartedPayload {
  year: number;
}

export interface PeriodAnnualClosedPayload {
  closing_id: number;
  closed_at:  string;
}

export interface PeriodAnnualReopenedPayload {
  closing_id:  number;
  reason:      string;
  reopened_at: string;
}
