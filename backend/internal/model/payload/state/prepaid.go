package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
)

// PrepaidCreatedState 建立預付費用時所需的狀態
type PrepaidCreatedState struct {
	Ledger      *projection.LedgerAccount
	PrepaidID   int64                             // populated by PrepaidProjectionService.applyCreated; used internally by TransactionProjectionService
	PrepaidUUID string                            // populated by PrepaidProjectionService.applyCreated; event UUID of the created prepaid
	Transaction payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s PrepaidCreatedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }

// PrepaidAmortizedState 執行攤提時所需的狀態
type PrepaidAmortizedState struct {
	Prepaid     *projection.Prepaid
	Transaction payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s PrepaidAmortizedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }

// PrepaidDisposedState 提前終止預付費用時所需的狀態
type PrepaidDisposedState struct {
	Prepaid     *projection.Prepaid
	Transaction payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s PrepaidDisposedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }
