package enums

import "akatengu/internal/pkg/enumx"

// ─────────────────────────────────────────
// AssetType 投資資產的類型
// ─────────────────────────────────────────

// AssetType 投資資產的類型
type AssetType = enumx.EnumVal[asetTypeVal]
type asetTypeVal string

const (
	assetTypeStock asetTypeVal = "STOCK"
	assetTypeFund  asetTypeVal = "FUND"
	assetTypeGold  asetTypeVal = "GOLD"
	assetTypeFX    asetTypeVal = "FX"
)

var validAssetTypes = []asetTypeVal{
	assetTypeStock,
	assetTypeFund,
	assetTypeGold,
	assetTypeFX,
}

var (
	AssetTypeStock = enumx.Must(string(assetTypeStock), validAssetTypes)
	AssetTypeFund  = enumx.Must(string(assetTypeFund), validAssetTypes)
	AssetTypeGold  = enumx.Must(string(assetTypeGold), validAssetTypes)
	AssetTypeFX    = enumx.Must(string(assetTypeFX), validAssetTypes)
)

func ParseAssetType(s string) (AssetType, error) {
	return enumx.New(s, validAssetTypes)
}

// ─────────────────────────────────────────
// CostMethod 計價成本的方法
// ─────────────────────────────────────────

// CostMethod 計價成本的方法
type CostMethod = enumx.EnumVal[costMethodVal]

type costMethodVal string

const (
	costMethodAvg  costMethodVal = "AVG"
	costMethodFIFO costMethodVal = "FIFO"
)

var validCostMethods = []costMethodVal{
	costMethodAvg,
	costMethodFIFO,
}

var (
	CostMethodAvg  = enumx.Must(string(costMethodAvg), validCostMethods)
	CostMethodFIFO = enumx.Must(string(costMethodFIFO), validCostMethods)
)

func ParseCostMethod(s string) (CostMethod, error) {
	return enumx.New(s, validCostMethods)
}

// ─────────────────────────────────────────
// MovementType 投資變動的方向
// ─────────────────────────────────────────

// MovementType 投資變動的方向
type MovementType = enumx.EnumVal[movementTypeVal]
type movementTypeVal string

const (
	movementTypeBuy      movementTypeVal = "BUY"
	movementTypeSell     movementTypeVal = "SELL"
	movementTypeDividend movementTypeVal = "DIVIDEND"
	movementTypeSplit    movementTypeVal = "SPLIT"
	movementTypeConvert  movementTypeVal = "CONVERT"
)

var validMovementTypes = []movementTypeVal{
	movementTypeBuy,
	movementTypeSell,
	movementTypeDividend,
	movementTypeSplit,
	movementTypeConvert,
}

var (
	MovementTypeBuy      = enumx.Must(string(movementTypeBuy), validMovementTypes)
	MovementTypeSell     = enumx.Must(string(movementTypeSell), validMovementTypes)
	MovementTypeDividend = enumx.Must(string(movementTypeDividend), validMovementTypes)
	MovementTypeSplit    = enumx.Must(string(movementTypeSplit), validMovementTypes)
	MovementTypeConvert  = enumx.Must(string(movementTypeConvert), validMovementTypes)
)

func ParseMovementType(s string) (MovementType, error) { return enumx.New(s, validMovementTypes) }

// ─────────────────────────────────────────
// LotStatus 庫存的狀態
// ─────────────────────────────────────────

// LotStatus 庫存的狀態
type LotStatus = enumx.EnumVal[lotStatusVal]
type lotStatusVal string

const (
	lotStatusOpen    lotStatusVal = "OPEN"
	lotStatusPartial lotStatusVal = "PARTIAL"
	lotStatusClose   lotStatusVal = "CLOSED"
)

var lotStatuses = []lotStatusVal{
	lotStatusOpen,
	lotStatusPartial,
	lotStatusClose,
}

var (
	LotStatusOpen    = enumx.Must(string(lotStatusOpen), lotStatuses)
	LotStatusPartial = enumx.Must(string(lotStatusPartial), lotStatuses)
	LotStatusClose   = enumx.Must(string(lotStatusClose), lotStatuses)
)

func ParseLotStatus(s string) (LotStatus, error) { return enumx.New(s, lotStatuses) }

// ─────────────────────────────────────────
// RateSource
// ─────────────────────────────────────────

type RateSource = enumx.EnumVal[rateSourceVal]
type rateSourceVal string

const (
	rateSourceManual rateSourceVal = "MANUAL"
	rateSourceImport rateSourceVal = "IMPORT"
)

var validRateSources = []rateSourceVal{
	rateSourceManual,
	rateSourceImport,
}

var (
	RateSourceManual = enumx.Must(string(rateSourceManual), validRateSources)
	RateSourceImport = enumx.Must(string(rateSourceImport), validRateSources)
)

func ParseRateSource(s string) (RateSource, error) { return enumx.New(s, validRateSources) }
