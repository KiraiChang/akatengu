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

type ClosingStatus = enumx.EnumVal[closingStatusVal]

type closingStatusVal string

var validClosingStatuses = []closingStatusVal{
	closingStatusOpen,
	closingStatusClosed,
	closingStatusReopened,
}

const (
	closingStatusOpen     closingStatusVal = "OPEN"
	closingStatusClosed   closingStatusVal = "CLOSED"
	closingStatusReopened closingStatusVal = "REOPENED"
)

var (
	ClosingStatusOpen     = enumx.Must(string(closingStatusOpen), validClosingStatuses)
	ClosingStatusClosed   = enumx.Must(string(closingStatusClosed), validClosingStatuses)
	ClosingStatusReopened = enumx.Must(string(closingStatusReopened), validClosingStatuses)
)

func ParseClosingStatus(s string) (ClosingStatus, error) {
	return enumx.New(s, validClosingStatuses)
}
