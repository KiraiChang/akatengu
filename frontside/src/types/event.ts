export interface AppendEventCmd<T> {
  aggregate_type:   string;
  aggregate_id:     string;
  expected_version: number;
  event_type:       string;
  payload:          T;
  metadata:         Record<string, unknown>;
}
