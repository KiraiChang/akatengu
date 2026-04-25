package payload

type PeriodMonthStartedPayload struct {
	PeriodStart string `json:"period_start"`
}

func (p PeriodMonthStartedPayload) Validate() error {
	var errs []string
	if p.PeriodStart == "" {
		errs = append(errs, "period_start is required")
	}

	return joinErrors(errs)
}

type PeriodMonthClosedPayload struct {
	ClosingId int64  `json:"closing_id"`
	ClosedAt  string `json:"closed_at"`
}

func (p PeriodMonthClosedPayload) Validate() error {
	var errs []string

	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ClosedAt == "" {
		errs = append(errs, "closed_at is required")
	}

	return joinErrors(errs)
}

type PeriodMonthReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}

func (p PeriodMonthReopenedPayload) Validate() error {
	var errs []string
	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ReopenedAt == "" {
		errs = append(errs, "reopened_at is required")
	}

	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}

type PeriodAnnualStartedPayload struct {
	Year int `json:"year"`
}

func (p PeriodAnnualStartedPayload) Validate() error {
	var errs []string
	if p.Year == 0 {
		errs = append(errs, "year is required")
	}
	return joinErrors(errs)
}

type PeriodAnnualClosedPayload struct {
	ClosingId int64  `json:"closing_id"`
	ClosedAt  string `json:"closed_at"`
}

func (p PeriodAnnualClosedPayload) Validate() error {
	var errs []string
	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ClosedAt == "" {
		errs = append(errs, "closed_at is required")
	}
	return joinErrors(errs)
}

type PeriodAnnualReopenedPayload struct {
	ClosingId  int64  `json:"closing_id"`
	Reason     string `json:"reason"`
	ReopenedAt string `json:"reopened_at"`
}

func (p PeriodAnnualReopenedPayload) Validate() error {
	var errs []string
	if p.ClosingId == 0 {
		errs = append(errs, "closing_id is required")
	}

	if p.ReopenedAt == "" {
		errs = append(errs, "reopened_at is required")
	}

	if p.Reason == "" {
		errs = append(errs, "reason is required")
	}

	return joinErrors(errs)
}
