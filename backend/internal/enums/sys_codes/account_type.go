package sys_codes

import "akatengu/internal/pkg/enumx"

type SysAccount = enumx.Enum[sysAccountVal]

//enumx:enum
type sysAccountVal string

const (
	// Asset

	SysAccountAssetPrepaidInterest sysAccountVal = "SYS:ASSET:PREPAID_INTEREST"

	// Expense

	SysAccountExpenseLoanInterestExpense       sysAccountVal = "SYS:EXPENSE:LOAN:INTEREST"
	SysAccountExpenseCreditCardInterestExpense sysAccountVal = "SYS:EXPENSE:CREDIT_CARD:INTEREST"

	// Equity

	SysAccountEquityCloseNetIncome sysAccountVal = "SYS:EQUITY:CLOSE_NET_INCOME"
	SysAccountEquityEquityOpening  sysAccountVal = "SYS:EQUITY:EQUITY_OPENING"
)
