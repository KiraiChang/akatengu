package enums

import "akatengu/internal/pkg/enumx"

type AggregateType = enumx.EnumVal[aggregateTypeVal]
type aggregateTypeVal string

const (
	aggregateTransaction    aggregateTypeVal = "TRANSACTION"
	aggregateInstallment    aggregateTypeVal = "INSTALLMENT"
	aggregateReconciliation aggregateTypeVal = "RECONCILIATION"
	aggregateAccount        aggregateTypeVal = "ACCOUNT"
	aggregateInvestment     aggregateTypeVal = "INVESTMENT"
	aggregateClose          aggregateTypeVal = "CLOSE"
)

var validAggregateTypes = []aggregateTypeVal{
	aggregateTransaction,
	aggregateInstallment,
	aggregateReconciliation,
	aggregateAccount,
	aggregateInvestment,
	aggregateClose,
}

var (
	AggregateTransaction    = enumx.Must(string(aggregateReconciliation), validAggregateTypes)
	AggregateInstallment    = enumx.Must(string(aggregateInstallment), validAggregateTypes)
	AggregateReconciliation = enumx.Must(string(aggregateReconciliation), validAggregateTypes)
	AggregateAccount        = enumx.Must(string(aggregateAccount), validAggregateTypes)
	AggreagateInvestment    = enumx.Must(string(aggregateInvestment), validAggregateTypes)
	AggregateClose          = enumx.Must(string(aggregateClose), validAggregateTypes)
)

func ParseAggregateType(s string) (AggregateType, error) {
	return enumx.New(s, validAggregateTypes)
}
