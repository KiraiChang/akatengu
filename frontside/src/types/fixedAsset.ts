export type DepreciationMethod = 'STRAIGHT_LINE';
export type AssetPaymentType   = 'CASH' | 'LEASE';
export type FixedAssetStatus   = 'ACTIVE' | 'DISPOSED';

export interface FixedAsset {
  id:                              number;
  merchant_id:                     number;
  txn_id:                          number | null;
  name:                            string;
  asset_account_id:                string;
  accum_depreciation_account_id:   string;
  depreciation_expense_account_id: string;
  cost:                            string;
  residual_value:                  string;
  useful_life_months:              number;
  depreciation_method:             DepreciationMethod;
  payment_type:                    AssetPaymentType;
  total_depreciated:               string;
  depreciated_periods:             number;
  purchase_date:                   string;
  disposal_date:                   string | null;
  status:                          FixedAssetStatus;
  updated_by:                      string | null;
  updated_at:                      string | null;
  version:                         number;
}

export interface FixedAssetDepreciation {
  id:          number;
  merchant_id: number;
  asset_id:    number;
  txn_id:      number;
  period_date: string;
  amount:      string;
  updated_by:  string | null;
  updated_at:  string | null;
}

export interface AssetPurchasedPayload {
  name:                            string;
  asset_account_id:                string;
  accum_depreciation_account_id:   string;
  depreciation_expense_account_id: string;
  cost:                            string;
  residual_value:                  string;
  useful_life_months:              number;
  payment_type:                    AssetPaymentType;
  ledger_id:                       number | null;
  liability_account_id:            string;
  purchase_date:                   string;
  memo:                            string;
  note:                            string;
}

export interface AssetDepreciatedPayload {
  asset_id:    number;
  period_date: string;
}

export interface AssetDisposedPayload {
  asset_id:           number;
  disposal_date:      string;
  proceeds:           string;
  proceeds_ledger_id: number | null;
  gain_account_id:    string;
  loss_account_id:    string;
  memo:               string;
}
