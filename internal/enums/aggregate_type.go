package enums

//go:generate go run ../../cmd/enumx-gen
import (
	"akatengu/internal/pkg/enumx"
)

type AggregateType = enumx.Enum[aggregateTypeVal]

//enumx:enum
type aggregateTypeVal string

const (
	AggregateTransaction aggregateTypeVal = "TRANSACTION"
	AggregateAccount     aggregateTypeVal = "ACCOUNT"
)
