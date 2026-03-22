package enums

import (
	"akatengu/internal/pkg/enumx"
)

// ─────────────────────────────────────────
// AccountType 會計科目類別
// ─────────────────────────────────────────

// AccountType 會計科目類別
type AccountType = enumx.EnumVal[accountTypeVal]
type accountTypeVal string

const (
	accountAsset     accountTypeVal = "ASSET"
	accountLiability accountTypeVal = "LIABILITY"
	accountEquity    accountTypeVal = "EQUITY"
	accountIncome    accountTypeVal = "INCOME"
	accountExpense   accountTypeVal = "EXPENSE"
)

var validAccountTypes = []accountTypeVal{
	accountAsset, accountLiability, accountEquity,
	accountIncome, accountExpense,
}

var (
	AccountAsset     = enumx.Must(string(accountAsset), validAccountTypes)
	AccountLiability = enumx.Must(string(accountLiability), validAccountTypes)
	AccountEquity    = enumx.Must(string(accountEquity), validAccountTypes)
	AccountIncome    = enumx.Must(string(accountIncome), validAccountTypes)
	AccountExpense   = enumx.Must(string(accountExpense), validAccountTypes)
)

func ParseAccountType(s string) (AccountType, error) {
	return enumx.New(s, validAccountTypes)
}

// ─────────────────────────────────────────
// NormalBalance 科目正常的借貸方向
// ─────────────────────────────────────────

// NormalBalance 科目正常的借貸方向
type NormalBalance = enumx.EnumVal[normalBalanceVal]
type normalBalanceVal string

const (
	balanceDebit  normalBalanceVal = "DEBIT"
	balanceCredit normalBalanceVal = "CREDIT"
)

var validNormalBalances = []normalBalanceVal{balanceDebit, balanceCredit}

var (
	BalanceDebit  = enumx.Must(string(balanceDebit), validNormalBalances)
	BalanceCredit = enumx.Must(string(balanceCredit), validNormalBalances)
)

func ParseNormalBalance(s string) (NormalBalance, error) {
	return enumx.New(s, validNormalBalances)
}

// ─────────────────────────────────────────
// TransactionStatus 會計分錄的狀態
// ─────────────────────────────────────────

// TransactionStatus 會計分錄的狀態
type TransactionStatus = enumx.EnumVal[transactionStatusVal]
type transactionStatusVal string

const (
	transactionStatusActive    transactionStatusVal = "ACTIVE"
	transactionStatusCorrected transactionStatusVal = "CORRECTED"
	transactionStatusVoided    transactionStatusVal = "VOIDED"
)

var validTransactionStatuses = []transactionStatusVal{
	transactionStatusActive,
	transactionStatusCorrected,
	transactionStatusVoided,
}

var (
	TransactionStatusActive    = enumx.Must(string(transactionStatusActive), validTransactionStatuses)
	TransactionStatusCorrected = enumx.Must(string(transactionStatusCorrected), validTransactionStatuses)
	TransactionStatusVoided    = enumx.Must(string(transactionStatusVoided), validTransactionStatuses)
)

func ParseTransactionStatus(s string) (TransactionStatus, error) {
	return enumx.New(s, validTransactionStatuses)
}
