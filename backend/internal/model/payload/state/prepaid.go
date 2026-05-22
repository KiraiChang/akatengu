package state

import "akatengu/internal/model/db/projection"

// PrepaidCreatedState 建立預付費用時所需的狀態
type PrepaidCreatedState struct {
	Ledger    *projection.LedgerAccount
	PrepaidID int64 // populated by PrepaidProjectionService.applyCreated
}

// PrepaidAmortizedState 執行攤提時所需的狀態
type PrepaidAmortizedState struct {
	Prepaid *projection.Prepaid
}

// PrepaidDisposedState 提前終止預付費用時所需的狀態
type PrepaidDisposedState struct {
	Prepaid *projection.Prepaid
}
