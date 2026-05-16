package enums

import (
	"akatengu/internal/pkg/enumx"
)

// ─────────────────────────────────────────
// AccountType 會計科目類別
// ─────────────────────────────────────────

// AccountType 會計科目類別
type AccountType = enumx.Enum[accountTypeVal]

//enumx:enum
type accountTypeVal string

const (
	AccountAsset     accountTypeVal = "ASSET"
	AccountLiability accountTypeVal = "LIABILITY"
	AccountEquity    accountTypeVal = "EQUITY"
	AccountIncome    accountTypeVal = "INCOME"
	AccountExpense   accountTypeVal = "EXPENSE"
)

// ─────────────────────────────────────────
// NormalBalance 科目正常的借貸方向
// ─────────────────────────────────────────

// NormalBalance 科目正常的借貸方向
type NormalBalance = enumx.Enum[normalBalanceVal]

//enumx:enum
type normalBalanceVal string

const (
	BalanceDebit  normalBalanceVal = "DEBIT"
	BalanceCredit normalBalanceVal = "CREDIT"
)

// ─────────────────────────────────────────
// TransactionStatus 會計分錄的狀態
// ─────────────────────────────────────────

// TransactionStatus 會計分錄的狀態
type TransactionStatus = enumx.Enum[transactionStatusVal]

//enumx:enum
type transactionStatusVal string

const (
	TransactionStatusActive    transactionStatusVal = "ACTIVE"
	TransactionStatusCorrected transactionStatusVal = "CORRECTED"
	TransactionStatusVoided    transactionStatusVal = "VOIDED"
	TransactionStatusVoidRef   transactionStatusVal = "VOID_REF"
)

type LedgerAccountType = enumx.Enum[ledgerAccountTypeVal]

//enumx:enum
type ledgerAccountTypeVal string

const (
	LedgerAccountTypeBankAccount ledgerAccountTypeVal = "BANK_ACCOUNT"
	LedgerAccountTypeLoan        ledgerAccountTypeVal = "LOAN"
	LedgerAccountTypeCreditCard  ledgerAccountTypeVal = "CREDIT_CARD"
)

// ─────────────────────────────────────────
// CashFlowCategory 現金流量表活動分類
// ─────────────────────────────────────────

// CashFlowCategory 現金流量表活動分類
type CashFlowCategory = enumx.Enum[cashFlowCategoryVal]

//enumx:enum
type cashFlowCategoryVal string

const (
	CashFlowCategoryCash      cashFlowCategoryVal = "CASH"
	CashFlowCategoryOperating cashFlowCategoryVal = "OPERATING"
	CashFlowCategoryInvesting cashFlowCategoryVal = "INVESTING"
	CashFlowCategoryFinancing cashFlowCategoryVal = "FINANCING"
)
