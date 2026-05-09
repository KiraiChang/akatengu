package enums

import "akatengu/internal/pkg/enumx"

// ─────────────────────────────────────────
// AssetType 投資資產的類型
// ─────────────────────────────────────────

// AssetType 投資資產的類型
type AssetType = enumx.Enum[assetTypeVal]

//enumx:enum
type assetTypeVal string

const (
	AssetTypeStock assetTypeVal = "STOCK"
	AssetTypeFund  assetTypeVal = "FUND"
	AssetTypeGold  assetTypeVal = "GOLD"
	AssetTypeFX    assetTypeVal = "FX"
)

// ─────────────────────────────────────────
// CostMethod 計價成本的方法
// ─────────────────────────────────────────

// CostMethod 計價成本的方法
type CostMethod = enumx.Enum[costMethodVal]

//enumx:enum
type costMethodVal string

const (
	CostMethodAvg  costMethodVal = "AVG"
	CostMethodFIFO costMethodVal = "FIFO"
)

// ─────────────────────────────────────────
// MovementType 投資變動的方向
// ─────────────────────────────────────────

// MovementType 投資變動的方向
type MovementType = enumx.Enum[movementTypeVal]

//enumx:enum
type movementTypeVal string

const (
	MovementTypeBuy      movementTypeVal = "BUY"
	MovementTypeSell     movementTypeVal = "SELL"
	MovementTypeDividend movementTypeVal = "DIVIDEND"
	MovementTypeSplit    movementTypeVal = "SPLIT"
	MovementTypeConvert  movementTypeVal = "CONVERT"
	MovementTypeMark     movementTypeVal = "MARK"
)

// ─────────────────────────────────────────
// LotStatus 庫存的狀態
// ─────────────────────────────────────────

// LotStatus 庫存的狀態
type LotStatus = enumx.Enum[lotStatusVal]

//enumx:enum
type lotStatusVal string

const (
	LotStatusOpen    lotStatusVal = "OPEN"
	LotStatusPartial lotStatusVal = "PARTIAL"
	LotStatusClose   lotStatusVal = "CLOSED"
)

// ─────────────────────────────────────────
// IFRSCategory IFRS 9 分類
// ─────────────────────────────────────────

// IFRSCategory IFRS 9 金融工具分類
type IFRSCategory = enumx.Enum[ifrsCategoryVal]

//enumx:enum
type ifrsCategoryVal string

const (
	IFRSCategoryFVTPL ifrsCategoryVal = "FVTPL"
	IFRSCategoryFVOCI ifrsCategoryVal = "FVOCI"
	IFRSCategoryAC    ifrsCategoryVal = "AC"
)

// ─────────────────────────────────────────
// RateSource
// ─────────────────────────────────────────

type RateSource = enumx.Enum[rateSourceVal]

//enumx:enum
type rateSourceVal string

const (
	RateSourceManual rateSourceVal = "MANUAL"
	RateSourceImport rateSourceVal = "IMPORT"
)
