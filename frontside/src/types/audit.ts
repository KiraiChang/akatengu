export interface AggregateVersionAudit {
  aggregate_type:  string;
  aggregate_id:    string;
  merchant_id:     number;
  current_version: number;
}

export interface EventStoreAudit {
  event_id:          number;
  event_uuid:        string;
  occurred_at:       string;
  merchant_id:       number;
  aggregate_type:    string;
  aggregate_id:      string;
  aggregate_version: number;
  event_type:        string;
  payload:           unknown;
  metadata:          unknown;
  updated_by:        string | null;
}

export interface CheckpointAudit {
  projection_name: string;
  merchant_id:     number;
  last_event_id:   number;
  updated_at:      string | null;
}

export interface SnapshotAudit {
  snapshot_id:    number;
  merchant_id:    number;
  aggregate_type: string;
  aggregate_id:   string;
  at_version:     number;
  state:          unknown;
  created_at:     string;
}

export interface ExchangeRate {
  rate_id:   number;
  currency:  string;
  rate_date: string;
  rate_twd:  string;
  source:    string;
}
