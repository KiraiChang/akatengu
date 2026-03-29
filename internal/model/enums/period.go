package enums

import "akatengu/internal/pkg/enumx"

type PeriodType = enumx.EnumVal[periodTypeVal]

type periodTypeVal string

const (
	periodMonthly periodTypeVal = "MONTHLY"
	periodAnnual  periodTypeVal = "ANNUAL"
)

var validPeriodTypes = []periodTypeVal{
	periodMonthly,
	periodAnnual,
}

var (
	PeriodMonthly = enumx.Must(string(periodMonthly), validPeriodTypes)
	PeriodAnnual  = enumx.Must(string(periodAnnual), validPeriodTypes)
)

func ParsePeriodType(s string) (PeriodType, error) {
	return enumx.New(s, validPeriodTypes)
}

type PeriodTypeStatus = enumx.EnumVal[periodTypeStatusVal]

type periodTypeStatusVal string

var validPeriodTypeStatuses = []periodTypeStatusVal{
	periodTypeStatusOpen,
	periodTypeStatusClosed,
	periodTypeStatusReopened,
}

const (
	periodTypeStatusOpen     periodTypeStatusVal = "OPEN"
	periodTypeStatusClosed   periodTypeStatusVal = "CLOSED"
	periodTypeStatusReopened periodTypeStatusVal = "REOPENED"
)

var (
	PeriodTypeStatusOpen     = enumx.Must(string(periodTypeStatusOpen), validPeriodTypeStatuses)
	PeriodTypeStatusClosed   = enumx.Must(string(periodTypeStatusClosed), validPeriodTypeStatuses)
	PeriodTypeStatusReopened = enumx.Must(string(periodTypeStatusReopened), validPeriodTypeStatuses)
)

func ParsePeriodTypeStatus(s string) (PeriodTypeStatus, error) {
	return enumx.New(s, validPeriodTypeStatuses)
}
