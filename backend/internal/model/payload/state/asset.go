package state

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
)

// AssetPurchasedState 購入固定資產時所需的狀態
type AssetPurchasedState struct {
	Category    *projection.FixedAssetCategory    // populated by pipeline; provides account IDs
	Ledger      *projection.LedgerAccount         // CASH：付款帳戶；LEASE：nil
	AssetID     int64                             // populated by FixedAssetProjectionService.applyPurchased; used internally by TransactionProjectionService
	AssetUUID   string                            // populated by FixedAssetProjectionService.applyPurchased; event UUID of the created asset
	Transaction payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s AssetPurchasedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }

// AssetPurchasedWithInstallmentState 以分期購入固定資產時所需的狀態
type AssetPurchasedWithInstallmentState struct {
	Category                       *projection.FixedAssetCategory // populated by pipeline; provides account IDs
	AssetID                        int64
	AssetUUID                      string
	Installment                    *projection.Installment
	InstallmentPayments            []*projection.InstallmentPayment
	SysAccountAssetPrepaidInterest string
	Transaction                    payload.TransactionCreatedPayload
}

func (s AssetPurchasedWithInstallmentState) GetTransaction() payload.TransactionCreatedPayload {
	return s.Transaction
}

// AssetDepreciatedState 執行折舊時所需的狀態
type AssetDepreciatedState struct {
	Asset       *projection.FixedAsset
	Transaction payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s AssetDepreciatedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }

// AssetDisposedState 處分固定資產時所需的狀態
type AssetDisposedState struct {
	Asset          *projection.FixedAsset
	ProceedsLedger *projection.LedgerAccount        // 收款帳戶；無收款時為 nil
	Transaction    payload.TransactionCreatedPayload // populated by factory/pipeline; read by TransactionProjectionService and AccountBalanceRealtimeProjection
}

func (s AssetDisposedState) GetTransaction() payload.TransactionCreatedPayload { return s.Transaction }
