package sys_codes

import "akatengu/internal/pkg/enumx"

type SysAccount = enumx.Enum[sysAccountVal]

//enumx:enum
type sysAccountVal string

const (
	// Asset

	SysAccountAssetPrepaidInterest sysAccountVal = "SYS:ASSET:PREPAID_INTEREST"

	// Equity

	SysAccountEquityCloseNetIncome sysAccountVal = "SYS:EQUITY:CLOSE_NET_INCOME"
	SysAccountEquityEquityOpening  sysAccountVal = "SYS:EQUITY:EQUITY_OPENING"

	// INCOME

	SysAccountIncomeFVTPLStockNetIncome sysAccountVal = "SYS:INCOME:FVTPL_STOCK_NET_INCOME"
	SysAccountIncomeFVTPLFundNetIncome  sysAccountVal = "SYS:INCOME:FVTPL_FUND_NET_INCOME"
	SysAccountIncomeFVTPLGoldNetIncome  sysAccountVal = "SYS:INCOME:FVTPL_GOLD_NET_INCOME"
	SysAccountIncomeFVTPLFXNetIncome    sysAccountVal = "SYS:INCOME:FVTPL_FX_NET_INCOME"

	// Expense

	SysAccountExpenseLoanInterestExpense       sysAccountVal = "SYS:EXPENSE:LOAN:INTEREST"
	SysAccountExpenseCreditCardInterestExpense sysAccountVal = "SYS:EXPENSE:CREDIT_CARD:INTEREST"

	SysAccountExpenseFVTPLStockLoss sysAccountVal = "SYS:Expense:FVTPL_STOCK_LOSS"
	SysAccountExpenseFVTPLFundLoss  sysAccountVal = "SYS:Expense:FVTPL_FUND_LOSS"
	SysAccountExpenseFVTPLGoldLoss  sysAccountVal = "SYS:Expense:FVTPL_GOLD_LOSS"
	SysAccountExpenseFVTPLFXLoss    sysAccountVal = "SYS:Expense:FVTPL_FX_LOSS"

	SysAccountExpenseInvestmentStockFee sysAccountVal = "SYS:EXPENSE:INVESTMENT_STOCK_FEE"
	SysAccountExpenseInvestmentFundFee  sysAccountVal = "SYS:EXPENSE:INVESTMENT_FUND_FEE"
	SysAccountExpenseInvestmentGoldFee  sysAccountVal = "SYS:EXPENSE:INVESTMENT_GOLD_FEE"
	SysAccountExpenseInvestmentFXFee    sysAccountVal = "SYS:EXPENSE:INVESTMENT_FX_FEE"

	SysAccountExpenseInvestmentStockTax sysAccountVal = "SYS:EXPENSE:INVESTMENT_STOCK_TAX"
	SysAccountExpenseInvestmentFundTax  sysAccountVal = "SYS:EXPENSE:INVESTMENT_FUND_TAX"
	SysAccountExpenseInvestmentGoldTax  sysAccountVal = "SYS:EXPENSE:INVESTMENT_GOLD_TAX"
	SysAccountExpenseInvestmentFXTax    sysAccountVal = "SYS:EXPENSE:INVESTMENT_FX_TAX"

	// FVTPL 未實現評價利益（各 AssetType → 4204-xx）
	SysAccountIncomeFVTPLStockUnrealized sysAccountVal = "SYS:INCOME:FVTPL_STOCK_UNREALIZED" // 4204-01
	SysAccountIncomeFVTPLFundUnrealized  sysAccountVal = "SYS:INCOME:FVTPL_FUND_UNREALIZED"  // 4204-02
	SysAccountIncomeFVTPLGoldUnrealized  sysAccountVal = "SYS:INCOME:FVTPL_GOLD_UNREALIZED"  // 4204-03
	SysAccountIncomeFVTPLFXUnrealized    sysAccountVal = "SYS:INCOME:FVTPL_FX_UNREALIZED"    // 4204-04

	// FVTPL 公允價值變動損失（5402-21）
	SysAccountExpenseFVTPLUnrealizedLoss sysAccountVal = "SYS:EXPENSE:FVTPL_UNREALIZED_LOSS" // 5402-21

	// FVOCI 其他綜合損益（3102-02）
	SysAccountEquityOCI sysAccountVal = "SYS:EQUITY:OCI" // 3102-02
)
