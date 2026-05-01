export type AssetType  = 'STOCK' | 'FUND' | 'GOLD' | 'FX';
export type CostMethod = 'AVG' | 'FIFO';

export interface Investment {
  investment_id: number;
  account_id:    string;
  asset_type:    AssetType;
  currency:      string;
  symbol:        string;
  name:          string;
  cost_method:   CostMethod;
  is_active:     boolean;
  version:       number;
}

export interface InvestmentCreatedPayload {
  account_id:  string;
  asset_type:  AssetType;
  currency:    string;
  symbol:      string;
  name:        string;
  cost_method: CostMethod;
  is_active:   boolean;
}

export interface InvestmentUpdatedPayload extends InvestmentCreatedPayload {
  investment_id: number;
  version:       number;
}

export type LotStatus = 'OPEN' | 'PARTIAL' | 'CLOSED';

export interface InvestmentLot {
  lot_id:         number;
  investment_id:  number;
  movement_id:    number;
  acquired_date:  string;
  transaction_id: number | null;
  quantity:      string;
  unit_cost:      string;
  total_cost:     string;
  remaining_qty:  string;
  status:        LotStatus;
}

export interface InvestmentPosition {
  id:            number;
  investment_id:  number;
  total_quantity: string;
  total_cost:     string;
  avg_cost:       string;
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
  lot_id:              number | null;
  movement_id:         number | null;
  quantity:            string;
  cost_basis:          string;
  sale_proceeds:       string;
  capital_gain:        string;
  holding_period_days: number | null;
  disposal_date:       string;
}