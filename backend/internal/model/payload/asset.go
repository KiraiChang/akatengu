package payload

import (
	"akatengu/internal/enums"

	"github.com/shopspring/decimal"
)

// AssetPurchasedPayload 購入固定資產
type AssetPurchasedPayload struct {
	Name                         string                 `json:"name"`
	AssetAccountID               string                 `json:"asset_account_id"`                // e.g. 1201-03 車輛
	AccumDepreciationAccountID   string                 `json:"accum_depreciation_account_id"`   // e.g. 1201-99
	DepreciationExpenseAccountID string                 `json:"depreciation_expense_account_id"` // e.g. 5501-02
	Cost                         decimal.Decimal        `json:"cost"`
	ResidualValue                decimal.Decimal        `json:"residual_value"`      // 殘值，預設 0
	UsefulLifeMonths             int64                  `json:"useful_life_months"`  // 耐用年限（月）
	PaymentType                  enums.AssetPaymentType `json:"payment_type"`        // CASH | LEASE
	LedgerID                     *int64                 `json:"ledger_id"`           // CASH：付款帳戶 ledger；LEASE：nil
	LiabilityAccountID           string                 `json:"liability_account_id"` // LEASE：租賃負債科目（如 2202-01）；CASH：""
	PurchaseDate                 string                 `json:"purchase_date"`       // YYYY-MM-DD
	Memo                         string                 `json:"memo,omitempty"`
	Note                         string                 `json:"note,omitempty"`
}

func (p AssetPurchasedPayload) Validate() error {
	var errs []string
	if p.Name == "" {
		errs = append(errs, "name is required")
	}
	if p.AssetAccountID == "" {
		errs = append(errs, "asset_account_id is required")
	}
	if p.AccumDepreciationAccountID == "" {
		errs = append(errs, "accum_depreciation_account_id is required")
	}
	if p.DepreciationExpenseAccountID == "" {
		errs = append(errs, "depreciation_expense_account_id is required")
	}
	if p.Cost.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "cost must be greater than zero")
	}
	if p.ResidualValue.LessThan(decimal.Zero) {
		errs = append(errs, "residual_value must be >= 0")
	}
	if p.UsefulLifeMonths <= 0 {
		errs = append(errs, "useful_life_months must be greater than zero")
	}
	if !p.PaymentType.In(enums.AllAssetPaymentType()...) {
		errs = append(errs, "payment_type is required")
	}
	if p.PaymentType == enums.AssetPaymentTypeCash.Enum() && (p.LedgerID == nil || *p.LedgerID <= 0) {
		errs = append(errs, "ledger_id is required for CASH payment type")
	}
	if p.PaymentType == enums.AssetPaymentTypeLease.Enum() && p.LiabilityAccountID == "" {
		errs = append(errs, "liability_account_id is required for LEASE payment type")
	}
	if p.PurchaseDate == "" {
		errs = append(errs, "purchase_date is required")
	}
	return joinErrors(errs)
}

// AssetDepreciatedPayload 執行一期折舊
type AssetDepreciatedPayload struct {
	AssetID    int64  `json:"asset_id"`
	PeriodDate string `json:"period_date"` // YYYY-MM
}

func (p AssetDepreciatedPayload) Validate() error {
	var errs []string
	if p.AssetID <= 0 {
		errs = append(errs, "asset_id is required")
	}
	if p.PeriodDate == "" {
		errs = append(errs, "period_date is required")
	}
	return joinErrors(errs)
}

// AssetDisposedPayload 處分固定資產
type AssetDisposedPayload struct {
	AssetID              int64           `json:"asset_id"`
	DisposalDate         string          `json:"disposal_date"`          // YYYY-MM-DD
	Proceeds             decimal.Decimal `json:"proceeds"`               // 處分收益（現金流入），0 表示報廢無收入
	ProceedsLedgerID     *int64          `json:"proceeds_ledger_id"`     // 收款帳戶；proceeds=0 時可為 nil
	GainAccountID        string          `json:"gain_account_id"`        // 處分利得科目 e.g. 4205
	LossAccountID        string          `json:"loss_account_id"`        // 處分損失科目 e.g. 5601
	Memo                 string          `json:"memo,omitempty"`
}

func (p AssetDisposedPayload) Validate() error {
	var errs []string
	if p.AssetID <= 0 {
		errs = append(errs, "asset_id is required")
	}
	if p.DisposalDate == "" {
		errs = append(errs, "disposal_date is required")
	}
	if p.Proceeds.LessThan(decimal.Zero) {
		errs = append(errs, "proceeds must be >= 0")
	}
	if p.Proceeds.GreaterThan(decimal.Zero) && (p.ProceedsLedgerID == nil || *p.ProceedsLedgerID <= 0) {
		errs = append(errs, "proceeds_ledger_id is required when proceeds > 0")
	}
	if p.GainAccountID == "" {
		errs = append(errs, "gain_account_id is required")
	}
	if p.LossAccountID == "" {
		errs = append(errs, "loss_account_id is required")
	}
	return joinErrors(errs)
}
