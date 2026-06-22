package event

// TracingData Tracing 有關的資料
type TracingData struct {
	CorrelationID string `json:"correlation_id"`
	CausationID   string `json:"causation_id"`
	RequestID     string `json:"request_id"`
}

// ExecutionData Execution control
type ExecutionData struct {
	PartitionKey string `json:"partitionKey"`
	Priority     int    `json:"priority"`
	Source       string `json:"source"` // api / replay / process
}

// SchemaData 資料庫相關
type SchemaData struct {
	SchemaVersion int    `json:"schemaVersion"`
	Encoding      string `json:"encoding"`
}

// ObservabilityData 觀察用資料
type ObservabilityData struct {
	Timestamp int64  `json:"timestamp"`
	Host      string `json:"host"`
}

// Metadata Event 怎麼發生的、在哪裡發生的、要怎麼處理它
type Metadata struct {
	Tracing       TracingData       `json:"tracing"`
	Execution     ExecutionData     `json:"execution"`
	Schema        SchemaData        `json:"schema"`
	Observability ObservabilityData `json:"observability"`
}
