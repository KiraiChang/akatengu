export interface UpdateAccountRequest {
  account_id:     string;
  parent_id:      string | null;
  name:           string;
  type:           string;
  normal_balance: string;
  currency:       string;
  is_summary:     boolean;
  is_active:      boolean;
  note:           string | null;
  version:        number;
}

export interface CreateAccountRequest {
  account_id:     string;
  parent_id:      string | null;
  name:           string;
  type:           string;
  normal_balance: string;
  currency:       string;
  is_summary:     boolean;
  is_active:      boolean;
  note:           string | null;
}

export interface Account {
  account_id:     string;
  parent_id:      string | null;
  name:          string;
  type:          string;
  normal_balance: string;
  currency:      string;
  is_summary:     boolean;
  is_active:      boolean;
  has_child:      boolean;
  note:          string | null;
  version:       number;
}
