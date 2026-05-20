import type { Account } from './account';
import type { LedgerAccountType } from './ledger';

export type { LedgerAccountType };
export type AssetType = 'STOCK' | 'FUND' | 'GOLD' | 'FX';

export interface LedgerAccountTypeConfig {
  id:         number;
  merchant_id: number;
  type:       LedgerAccountType;
  account_id: string;
  updated_by: string | null;
  updated_at: string | null;
  version:    number;
}

export interface LedgerAccountTypeConfigResult extends LedgerAccountTypeConfig {
  descendants: Account[];
}

export interface AssetTypeAccountConfig {
  id:                          number;
  merchant_id:                 number;
  asset_type:                  AssetType;
  realized_gain_account_id:    string;
  realized_loss_account_id:    string;
  unrealized_gain_account_id:  string;
  unrealized_loss_account_id:  string;
  oci_account_id:              string | null;
  fee_account_id:              string;
  tax_account_id:              string;
  updated_by:                  string | null;
  updated_at:                  string | null;
  version:                     number;
}

export interface UpdateAssetTypePayload {
  realized_gain_account_id:    string;
  realized_loss_account_id:    string;
  unrealized_gain_account_id:  string;
  unrealized_loss_account_id:  string;
  oci_account_id:              string | null;
  fee_account_id:              string;
  tax_account_id:              string;
}
