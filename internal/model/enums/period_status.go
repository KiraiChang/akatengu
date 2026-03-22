package enums

import "akatengu/internal/pkg/enumx"

type PeriodStatus = enumx.EnumVal[periodStatusVal]
type periodStatusVal string

const (
	periodOpen       periodStatusVal = "OPEN"
	periodSoftClosed periodStatusVal = "SOFT_CLOSED"
	periodHardClosed periodStatusVal = "HARD_CLOSED"
)

var validPeriodStatuses = []periodStatusVal{
	periodOpen, periodSoftClosed, periodHardClosed,
}

var (
	PeriodOpen       = enumx.Must(string(periodOpen), validPeriodStatuses)
	PeriodSoftClosed = enumx.Must(string(periodSoftClosed), validPeriodStatuses)
	PeriodHardClosed = enumx.Must(string(periodHardClosed), validPeriodStatuses)
)

func ParsePeriodStatus(s string) (PeriodStatus, error) {
	return enumx.New(s, validPeriodStatuses)
}
