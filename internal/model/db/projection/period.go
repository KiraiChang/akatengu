package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

type PeriodClosing struct {
	ClosingId    int64                  `db:"closing_id"`
	PeriodType   enums.PeriodType       `db:"period_type"`
	PeriodStart  string                 `db:"period_start"`
	PeriodEnd    string                 `db:"period_end"`
	Status       enums.PeriodTypeStatus `db:"status"`
	OpeningTxnID *int64                 `db:"opening_txn_id"`
	ClosingTxnID *int64                 `db:"closing_txn_id"`
	Snapshot     *string                `db:"snapshot"`
	ClosedAt     *string                `db:"closed_at"`
	Note         *string                `db:"note"`
	ReopenAt     *string                `db:"reopen_at"`
	ReopenReason *string                `db:"reopen_reason"`
}

// SnapshotData 是存入 snapshot 欄位的 JSON 結構
type SnapshotData struct {
	GeneratedAt      string              `json:"generated_at"`
	TotalAssets      decimal.Decimal     `json:"total_assets"`
	TotalLiabilities decimal.Decimal     `json:"total_liabilities"`
	NetWorth         decimal.Decimal     `json:"net_worth"`
	Assets           []BalanceSheetEntry `json:"assets"`
	Liabilities      []BalanceSheetEntry `json:"liabilities"`
	Equity           []BalanceSheetEntry `json:"equity"`
}

type BalanceSheetEntry struct {
	AccountId string          `json:"account_id"`
	Name      string          `json:"name"`
	Balance   decimal.Decimal `json:"balance"`
}
