package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
)

// ─────────────────────────────────────────
// InstallmentCreatedState
// ─────────────────────────────────────────

type InstallmentCreatedState struct {
	Ledger                         *projection.LedgerAccount
	Installment                    *projection.Installment
	InstallmentPayments            []*projection.InstallmentPayment
	SysAccountAssetPrepaidInterest string
	Transaction                    payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s InstallmentCreatedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }

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
	Transaction                      payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s InstallmentPeriodPaidState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }
