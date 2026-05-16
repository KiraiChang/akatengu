export interface BalanceSheetRow {
  type:       string;
  account_id: string;
  name:       string;
  balance:    string;
  parent_id:  string | null;
  has_child:  boolean;
  depth:      number;
}

export interface BalanceSheet {
  report_date:       string;
  assets:            BalanceSheetRow[];
  liabilities:       BalanceSheetRow[];
  equity:            BalanceSheetRow[];
  total_assets:      string;
  total_liabilities: string;
  total_equity:      string;
  net_worth:         string;
}

export interface IncomeStatementRow {
  type:       string;
  account_id: string;
  name:       string;
  amount:     string;
  parent_id:  string | null;
  has_child:  boolean;
  depth:      number;
}

export interface IncomeStatement {
  start_date:     string;
  end_date:       string;
  income:         IncomeStatementRow[];
  expenses:       IncomeStatementRow[];
  total_income:   string;
  total_expenses: string;
  net_income:     string;
}

export interface CashFlowItem {
  account_id: string;
  name:       string;
  amount:     string;
}

export interface CashFlowSection {
  items: CashFlowItem[];
  total: string;
}

export interface OperatingActivities {
  net_income:  string;
  adjustments: CashFlowItem[];
  total:       string;
}

export interface EquityItem {
  account_id:    string;
  name:          string;
  is_summary:    boolean;
  is_virtual:    boolean;
  begin_balance: string;
  period_change: string;
  end_balance:   string;
}

export interface EquityStatement {
  start_date:          string;
  end_date:            string;
  items:               EquityItem[];
  net_income:          string;
  total_begin_balance: string;
  total_period_change: string;
  total_end_balance:   string;
}

export interface CashFlowStatement {
  start_date:           string;
  end_date:             string;
  operating_activities: OperatingActivities;
  investing_activities: CashFlowSection;
  financing_activities: CashFlowSection;
  net_change:           string;
  beginning_cash:       string;
  ending_cash:          string;
}
