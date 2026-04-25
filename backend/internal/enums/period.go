package enums

import "akatengu/internal/pkg/enumx"

// ─────────────────────────────────────────
// PeriodType
// ─────────────────────────────────────────

type PeriodType = enumx.Enum[periodTypeVal]

//enumx:enum
type periodTypeVal string

const (
	PeriodMonthly periodTypeVal = "MONTHLY"
	PeriodAnnual  periodTypeVal = "ANNUAL"
)

// ─────────────────────────────────────────
// PeriodTypeStatus
// ─────────────────────────────────────────

type PeriodTypeStatus = enumx.Enum[periodTypeStatusVal]

//enumx:enum
type periodTypeStatusVal string

const (
	PeriodTypeStatusOpen     periodTypeStatusVal = "OPEN"
	PeriodTypeStatusClosed   periodTypeStatusVal = "CLOSED"
	PeriodTypeStatusReopened periodTypeStatusVal = "REOPENED"
)
