package pdfparser

import (
	"io"

	"github.com/shopspring/decimal"
)

// LedgerHint 告知 parser 此次匯入關聯哪些帳本，以便自動分配交易歸屬。
type LedgerHint struct {
	LedgerUUID  string
	AccountType string // BANK_ACCOUNT | CREDIT_CARD
}

// ParsedRow 是解析後的單筆交易，LedgerUUID 已由 parser 根據 LedgerHints 決定。
type ParsedRow struct {
	LedgerUUID  string
	TxnDate     string
	Description string
	Debit       decimal.Decimal
	Credit      decimal.Decimal
	Balance     *decimal.Decimal
	ReferenceNo *string
}

// ParseOptions 解析選項（如密碼），與 PDF 內容無關、不寫入資料庫。
type ParseOptions struct {
	Password *string
}

// BankPDFParser 是各銀行 PDF 解析器的統一介面。
// ledgers 來自主檔的 ledger 清單，parser 依 PDF 結構將每筆 row 分配到正確 LedgerUUID。
type BankPDFParser interface {
	Parse(r io.Reader, ledgers []LedgerHint, opts ParseOptions) ([]ParsedRow, error)
}
