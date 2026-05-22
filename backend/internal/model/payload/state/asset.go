package state

import "akatengu/internal/model/db/projection"

// AssetPurchasedState 購入固定資產時所需的狀態
type AssetPurchasedState struct {
	Ledger  *projection.LedgerAccount // CASH：付款帳戶；LEASE：nil
	AssetID int64                     // populated by FixedAssetProjectionService.applyPurchased
}

// AssetDepreciatedState 執行折舊時所需的狀態
type AssetDepreciatedState struct {
	Asset *projection.FixedAsset
}

// AssetDisposedState 處分固定資產時所需的狀態
type AssetDisposedState struct {
	Asset            *projection.FixedAsset
	ProceedsLedger   *projection.LedgerAccount // 收款帳戶；無收款時為 nil
}
