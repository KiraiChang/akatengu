package state

import "akatengu/internal/model/db/projection"

// ─────────────────────────────────────────
// InstallmentCreatedState
// ─────────────────────────────────────────

type InstallmentCreatedState struct {
	Ledger                         *projection.LedgerAccount
	Installment                    *projection.Installment
	InstallmentPayments            []*projection.InstallmentPayment
	SysAccountAssetPrepaidInterest string
}

// ─────────────────────────────────────────
// InstallmentPeriodPaidState
// ─────────────────────────────────────────

type InstallmentPeriodPaidState struct {
	PaidLedger                       *projection.LedgerAccount // 支付分類帳
	Ledger                           *projection.LedgerAccount // 應付帳款分類帳
	Installment                      *projection.Installment
	InstallmentPayments              *projection.InstallmentPayment
	SysAccountExpenseInterestExpense string
	SysAccountAssetPrepaidInterest   string
}
