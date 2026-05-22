package enums

import "akatengu/internal/pkg/enumx"

// DepreciationMethod 折舊方法
type DepreciationMethod = enumx.Enum[depreciationMethodVal]

//enumx:enum
type depreciationMethodVal string

const (
	DepreciationMethodStraightLine depreciationMethodVal = "STRAIGHT_LINE"
)

// AssetPaymentType 資產取得付款方式
type AssetPaymentType = enumx.Enum[assetPaymentTypeVal]

//enumx:enum
type assetPaymentTypeVal string

const (
	AssetPaymentTypeCash  assetPaymentTypeVal = "CASH"
	AssetPaymentTypeLease assetPaymentTypeVal = "LEASE"
)

// PrepaidStatus 預付費用狀態
type PrepaidStatus = enumx.Enum[prepaidStatusVal]

//enumx:enum
type prepaidStatusVal string

const (
	PrepaidStatusActive    prepaidStatusVal = "ACTIVE"
	PrepaidStatusCompleted prepaidStatusVal = "COMPLETED"
	PrepaidStatusDisposed  prepaidStatusVal = "DISPOSED"
)

// FixedAssetStatus 固定資產狀態
type FixedAssetStatus = enumx.Enum[fixedAssetStatusVal]

//enumx:enum
type fixedAssetStatusVal string

const (
	FixedAssetStatusActive   fixedAssetStatusVal = "ACTIVE"
	FixedAssetStatusDisposed fixedAssetStatusVal = "DISPOSED"
)
