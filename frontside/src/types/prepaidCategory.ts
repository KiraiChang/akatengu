export interface PrepaidCategory {
  id:               number;
  category_uuid:    string;
  merchant_id:      number;
  name:             string;
  account_id:       string;
  expense_account_id: string;
  is_active:        boolean;
  updated_by:       string | null;
  updated_at:       string;
  version:          number;
}

export interface PrepaidCategoryCreatePayload {
  name:               string;
  account_id:         string;
  expense_account_id: string;
}

export interface PrepaidCategoryUpdatePayload {
  expected_version:   number;
  name:               string;
  account_id:         string;
  expense_account_id: string;
}
