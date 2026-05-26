export interface DashboardSummary {
  as_of_date:        string;
  month:             string;
  month_income:      string;
  month_expense:     string;
  total_assets:      string;
  total_liabilities: string;
  total_equity:      string;
  cash_balance:      string;
}

export interface MonthlyTrendItem {
  month:   string;
  income:  string;
  expense: string;
  net:     string;
}

export interface LedgerBalance {
  ledger_id:   number;
  name:        string;
  institution: string;
  type:        string;
  balance:     string;
}
