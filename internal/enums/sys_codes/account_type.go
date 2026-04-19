package sys_codes

import "akatengu/internal/pkg/enumx"

type SysAccount = enumx.Enum[sysAccountVal]

//enumx:enum
type sysAccountVal string

const (
	// Asset

	SysAccountAssetPrepaidInterest sysAccountVal = "SYS:ASSET:PREPAID_INTEREST"

	// Expense

	SysAccountExpenseInterestExpense sysAccountVal = "SYS:EXPENSE:INTEREST"
)
