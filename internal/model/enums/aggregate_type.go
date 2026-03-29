package enums

import "akatengu/internal/pkg/enumx"

type AggregateType = enumx.EnumVal[aggregateTypeVal]
type aggregateTypeVal string

const (
	aggregateTransaction aggregateTypeVal = "TRANSACTION"
	aggregateAccount     aggregateTypeVal = "ACCOUNT"
)

var validAggregateTypes = []aggregateTypeVal{
	aggregateTransaction,
	aggregateAccount,
}

var (
	AggregateTransaction = enumx.Must(string(aggregateTransaction), validAggregateTypes)
	AggregateAccount     = enumx.Must(string(aggregateAccount), validAggregateTypes)
)

func ParseAggregateType(s string) (AggregateType, error) {
	return enumx.New(s, validAggregateTypes)
}
