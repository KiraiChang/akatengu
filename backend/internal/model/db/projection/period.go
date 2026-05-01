package projection

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

//dbmap:sqlcdb=PeriodClosing
//dbmap:sqlcdb=InsertPeriodCloseParams
//dbmap:sqlcdb=GetPeriodPagedByTypeRow
type PeriodClosing struct {
	ClosingId    int64                  `db:"closing_id" json:"closing_id"`
	PeriodType   enums.PeriodType       `db:"period_type" json:"period_type"`
	PeriodStart  string                 `db:"period_start" json:"period_start"`
	PeriodEnd    string                 `db:"period_end" json:"period_end"`
	Status       enums.PeriodTypeStatus `db:"status" json:"status"`
	OpeningTxnID *int64                 `db:"opening_txn_id" json:"opening_txn_id"`
	ClosingTxnID *int64                 `db:"closing_txn_id" json:"closing_txn_id"`
	Snapshot     *string                `db:"snapshot" json:"snapshot"`
	ClosedAt     *string                `db:"closed_at" json:"closed_at"`
	Note         *string                `db:"note" json:"note"`
	ReopenAt     *string                `db:"reopen_at" json:"reopen_at"`
	ReopenReason *string                `db:"reopen_reason" json:"reopen_reason"`
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
