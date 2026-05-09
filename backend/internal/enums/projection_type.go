package enums

import "akatengu/internal/pkg/enumx"

// ─────────────────────────────────────────
// ProjectionType event 直接轉成 db 資料的類別
// ─────────────────────────────────────────

// ProjectionType event 直接轉成 db 資料的類別
type ProjectionType = enumx.Enum[projectionTypeValue]

//enumx:enum
type projectionTypeValue string

const (
	ProjectionTypeAccount     projectionTypeValue = "ACCOUNT"
	ProjectionTypeInstallment projectionTypeValue = "INSTALLMENT"
	ProjectionTypeInvestment  projectionTypeValue = "INVESTMENT"
	ProjectionTypePeriod      projectionTypeValue = "PERIOD"
	ProjectionTypeTransaction              projectionTypeValue = "TRANSACTION"
	ProjectionTypeAccountBalanceSnapshot   projectionTypeValue = "ACCOUNT_BALANCE_SNAPSHOT"
)
