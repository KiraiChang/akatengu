export type CashFlowCategory = 'CASH' | 'OPERATING' | 'INVESTING' | 'FINANCING';

export const CASH_FLOW_CATEGORIES: CashFlowCategory[] = ['CASH', 'OPERATING', 'INVESTING', 'FINANCING'];

export const CASH_FLOW_CATEGORY_LABELS: Record<CashFlowCategory, string> = {
  CASH:      '現金及約當現金',
  OPERATING: '營業活動',
  INVESTING: '投資活動',
  FINANCING: '籌資活動',
};

export interface UpdateAccountRequest {
  account_id:          string;
  parent_id:           string | null;
  name:                string;
  type:                string;
  normal_balance:      string;
  currency:            string;
  is_summary:          boolean;
  is_active:           boolean;
  note:                string | null;
  version:             number;
  cash_flow_category?: CashFlowCategory | null;
}

export interface CreateAccountRequest {
  account_id:          string;
  parent_id:           string | null;
  name:                string;
  type:                string;
  normal_balance:      string;
  currency:            string;
  is_summary:          boolean;
  is_active:           boolean;
  note:                string | null;
  cash_flow_category?: CashFlowCategory | null;
}

export interface AccountBalance {
  account_id:     string;
  name:           string;
  type:           string;
  normal_balance: string;
  debit_total:    string;
  credit_total:   string;
}

export interface Account {
  account_id:         string;
  parent_id:          string | null;
  name:               string;
  type:               string;
  normal_balance:     string;
  currency:           string;
  is_summary:         boolean;
  is_active:          boolean;
  has_child:          boolean;
  note:               string | null;
  version:            number;
  cash_flow_category: CashFlowCategory | null;
}
