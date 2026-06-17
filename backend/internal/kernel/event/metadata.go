package event

// TracingData Tracing 有關的資料
type TracingData struct {
	CorrelationID string
	CausationID   string
	RequestID     string
}

// ExecutionData Execution control
type ExecutionData struct {
	PartitionKey string
	Priority     int
	Source       string // api / replay / process
}

// SchemaData 資料庫相關
type SchemaData struct {
	SchemaVersion int
	Encoding      string
}

// ObservabilityData 觀察用資料
type ObservabilityData struct {
	Timestamp int64
	Host      string
}

// Metadata Event 怎麼發生的、在哪裡發生的、要怎麼處理它
type Metadata struct {
	Tracing       TracingData
	Execution     ExecutionData
	Schema        SchemaData
	Observability ObservabilityData
}
