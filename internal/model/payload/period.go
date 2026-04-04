package payload

type PeriodMonthStartedPayload struct {
	PeriodStart string `json:"period_start"`
}

type PeriodMonthClosedPayload struct {
	ClosingId int64  `json:"closing_id"`
	ClosedAt  string `json:"closed_at"`
}

type PeriodMonthReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}

type PeriodAnnualStartedPayload struct {
	Year int `json:"year"`
}

type PeriodAnnualClosedPayload struct {
	ClosingId int64  `json:"closing_id"`
	ClosedAt  string `json:"closed_at"`
}

type PeriodAnnualReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}
