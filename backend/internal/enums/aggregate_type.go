package enums

//go:generate go run ../../cmd/enumx-gen
import (
	"akatengu/internal/pkg/enumx"
)

type AggregateType = enumx.Enum[aggregateTypeVal]

//enumx:enum
type aggregateTypeVal string

const (
	AggregateTransaction        aggregateTypeVal = "TRANSACTION"
	AggregateAccount            aggregateTypeVal = "ACCOUNT"
	AggregateSysConfig          aggregateTypeVal = "SYS_CONFIG"
	AggregateAccountConfig      aggregateTypeVal = "ACCOUNT_CONFIG"
	AggregateFixedAssetCategory aggregateTypeVal = "FIXED_ASSET_CATEGORY"
	AggregatePrepaidCategory    aggregateTypeVal = "PREPAID_CATEGORY"
	AggregateBankCsvTemplate    aggregateTypeVal = "BANK_CSV_TEMPLATE"
	AggregateBankStatement      aggregateTypeVal = "BANK_STATEMENT"
	AggregateBankPdfTemplate    aggregateTypeVal = "BANK_PDF_TEMPLATE"
)
