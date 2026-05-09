export type AssetType    = 'STOCK' | 'FUND' | 'GOLD' | 'FX';
export type CostMethod   = 'AVG' | 'FIFO';
export type IFRSCategory = 'FVTPL' | 'FVOCI' | 'AC';

export interface Investment {
  investment_id: number;
  account_id:    string;
  asset_type:    AssetType;
  currency:      string;
  symbol:        string;
  name:          string;
  cost_method:   CostMethod;
  ifrs_category: IFRSCategory;
  is_active:     boolean;
  version:       number;
}

export interface InvestmentCreatedPayload {
  account_id:    string;
  asset_type:    AssetType;
  currency:      string;
  symbol:        string;
  name:          string;
  cost_method:   CostMethod;
  ifrs_category: IFRSCategory;
  is_active:     boolean;
}

export interface InvestmentUpdatedPayload extends InvestmentCreatedPayload {
  investment_id: number;
  version:       number;
}

export type LotStatus    = 'OPEN' | 'PARTIAL' | 'CLOSED';
export type MovementType = 'BUY' | 'SELL' | 'DIVIDEND' | 'SPLIT' | 'CONVERT';

export interface InvestmentLot {
  lot_id:               number;
  investment_id:        number;
  movement_id:          number;
  acquired_date:        string;
  txn_id:               number | null;
  quantity:             string;
  unit_cost:            string;
  total_cost:           string;
  remaining_qty:        string;
  status:               LotStatus;
  unrealized_unit_twd:  string;
}

export interface InvestmentPosition {
  id:               number;
  investment_id:    number;
  total_quantity:   string;
  total_cost:       string;
  avg_cost:         string;
  market_price_twd: string;
}

export interface InvestmentBoughtPayload {
  investment_id: number;
  date:          string;
  quantity:      string;
  unit_price:    string;
  exchange_rate: string;
  fee:           string;
  tax:           string;
  ledger_id:     number;
}

export interface InvestmentSoldPayload {
  investment_id: number;
  date:          string;
  quantity:      string;
  unit_price:    string;
  exchange_rate: string;
  fee:           string;
  tax:           string;
  ledger_id:     number;
}

export interface InvestmentLotDisposal {
  id:                  number;
  lot_id:              number;
  movement_id:         number;
  txn_id:              number | null;
  quantity:            string;
  cost_basis:          string;
  sale_proceeds:       string;
  capital_gain:        string;
  holding_period_days: number;
  disposal_date:       string;
}

export interface InvestmentMovement {
  movement_id:     number;
  investment_id:   number;
  event_id:        number;
  txn_id:          number | null;
  movement_type:   MovementType;
  movement_date:   string;
  quantity:        string;
  unit_price:      string;
  unit_price_twd:  string;
  exchange_rate:   string;
  fee:             string;
  tax:             string;
  realized_gain:   string | null;
  cost_basis:      string | null;
  gross_amount:    string | null;
  net_amount:      string | null;
  withholding_tax: string | null;
  SplitRatio:      string | null;
}