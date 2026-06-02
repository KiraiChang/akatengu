export interface FixedAssetCategory {
  id:                              number;
  category_uuid:                   string;
  merchant_id:                     number;
  name:                            string;
  asset_account_id:                string;
  accum_depreciation_account_id:   string;
  depreciation_expense_account_id: string;
  is_active:                       boolean;
  updated_by:                      string | null;
  updated_at:                      string;
  version:                         number;
}

export interface FixedAssetCategoryCreatePayload {
  name:                            string;
  asset_account_id:                string;
  accum_depreciation_account_id:   string;
  depreciation_expense_account_id: string;
}

export interface FixedAssetCategoryUpdatePayload {
  expected_version:                number;
  name:                            string;
  asset_account_id:                string;
  accum_depreciation_account_id:   string;
  depreciation_expense_account_id: string;
}
