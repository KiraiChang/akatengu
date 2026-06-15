package payload

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"fmt"

	"github.com/shopspring/decimal"
)

// AssetPurchasedPayload 購入固定資產
type AssetPurchasedPayload struct {
	CategoryUUID     string                 `json:"category_uuid"`        // 固定資產類別 UUID
	Name             string                 `json:"name"`
	Cost             decimal.Decimal        `json:"cost"`
	ResidualValue    decimal.Decimal        `json:"residual_value"`       // 殘值，預設 0
	UsefulLifeMonths int64                  `json:"useful_life_months"`   // 耐用年限（月）
	PaymentType      enums.AssetPaymentType `json:"payment_type"`         // CASH | LEASE
	LedgerUUID       *string                `json:"ledger_uuid"`          // CASH：付款帳戶 ledger_uuid；LEASE：nil
	LiabilityAccountID string               `json:"liability_account_id"` // LEASE：租賃負債科目（如 2202-01）；CASH：""
	PurchaseDate     string                 `json:"purchase_date"`        // YYYY-MM-DD
	Memo             string                 `json:"memo,omitempty"`
	Note             string                 `json:"note,omitempty"`
}

func (p AssetPurchasedPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
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
	if p.PaymentType == enums.AssetPaymentTypeCash.Enum() && (p.LedgerUUID == nil || *p.LedgerUUID == "") {
		errs = append(errs, "ledger_uuid is required for CASH payment type")
	}
	if p.PaymentType == enums.AssetPaymentTypeLease.Enum() && p.LiabilityAccountID == "" {
		errs = append(errs, "liability_account_id is required for LEASE payment type")
	}
	if p.PurchaseDate == "" {
		errs = append(errs, "purchase_date is required")
	}
	return joinErrors(errs)
}

// AssetPurchasedWithInstallmentPayload 以信用卡分期購入固定資產
type AssetPurchasedWithInstallmentPayload struct {
	CategoryUUID     string                  `json:"category_uuid"` // 固定資產類別 UUID
	Name             string                  `json:"name"`
	Cost             decimal.Decimal         `json:"cost"`
	ResidualValue    decimal.Decimal         `json:"residual_value"`
	UsefulLifeMonths int64                   `json:"useful_life_months"`
	PurchaseDate     string                  `json:"purchase_date"`
	Memo             string                  `json:"memo,omitempty"`
	Note             string                  `json:"note,omitempty"`
	Installment      InstallmentTermsPayload `json:"installment"`
}

func (p AssetPurchasedWithInstallmentPayload) Validate() error {
	var errs []string
	if p.CategoryUUID == "" {
		errs = append(errs, "category_uuid is required")
	}
	if p.Name == "" {
		errs = append(errs, "name is required")
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
	if p.PurchaseDate == "" {
		errs = append(errs, "purchase_date is required")
	}
	if err := p.Installment.Validate(); err != nil {
		errs = append(errs, err.Error())
	}
	return joinErrors(errs)
}

// ToInstallmentPayload 組合完整的 InstallmentCreatedPayload，以資產科目與成本填入 AccountId / Amount。
// assetAccountID 從類別設定取得，由 Pipeline 傳入。
func (p AssetPurchasedWithInstallmentPayload) ToInstallmentPayload(assetAccountID string) InstallmentCreatedPayload {
	return InstallmentCreatedPayload{
		Amount:           p.Cost,
		InstallmentCount: p.Installment.InstallmentCount,
		StartDate:        p.Installment.StartDate,
		InterestType:     p.Installment.InterestType,
		AnnualRate:       p.Installment.AnnualRate,
		AccountId:        assetAccountID,
		LedgerUuid:       p.Installment.LedgerUuid,
		Memo:             p.Installment.Memo,
		Note:             p.Installment.Note,
	}
}

// BuildAssetPurchasedWithInstallmentTransaction 組合固定資產分期購入的會計分錄。
func BuildAssetPurchasedWithInstallmentTransaction(p AssetPurchasedWithInstallmentPayload, category *projection.FixedAssetCategory, ledger *projection.LedgerAccount, sysAccountAssetPrepaidInterest string, payments []*projection.InstallmentPayment) (TransactionCreatedPayload, error) {
	ip := p.ToInstallmentPayload(category.AssetAccountID)
	var entries []TransactionEntryPayload
	switch ip.InterestType.Val() {
	case enums.InterestTypeFree:
		entries = []TransactionEntryPayload{
			{AccountId: category.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledger.LedgerId, Debit: decimal.Zero, Credit: p.Cost},
		}
	case enums.InterestTypeFixedRate:
		interest := decimal.Zero
		for _, pmt := range payments {
			interest = interest.Add(pmt.Interest)
		}
		entries = []TransactionEntryPayload{
			{AccountId: category.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero},
			{AccountId: sysAccountAssetPrepaidInterest, Debit: interest, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledger.LedgerId, Debit: decimal.Zero, Credit: p.Cost.Add(interest)},
		}
	default:
		return TransactionCreatedPayload{}, fmt.Errorf("invalid interest type: %s", ip.InterestType.Val())
	}
	return TransactionCreatedPayload{
		TransactionDate: p.PurchaseDate,
		Description:     fmt.Sprintf("固定資產分期購入 %s", p.Name),
		Currency:        "TWD",
		Entries:         entries,
	}, nil
}

// AssetDepreciatedPayload 執行一期折舊
type AssetDepreciatedPayload struct {
	AssetUUID  string `json:"asset_uuid"`
	PeriodDate string `json:"period_date"` // YYYY-MM
}

func (p AssetDepreciatedPayload) Validate() error {
	var errs []string
	if p.AssetUUID == "" {
		errs = append(errs, "asset_uuid is required")
	}
	if p.PeriodDate == "" {
		errs = append(errs, "period_date is required")
	}
	return joinErrors(errs)
}

// AssetDisposedPayload 處分固定資產
type AssetDisposedPayload struct {
	AssetUUID            string          `json:"asset_uuid"`
	DisposalDate         string          `json:"disposal_date"`              // YYYY-MM-DD
	Proceeds             decimal.Decimal `json:"proceeds"`                   // 處分收益（現金流入），0 表示報廢無收入
	ProceedsLedgerUUID   *string         `json:"proceeds_ledger_uuid"`       // 收款帳戶；proceeds=0 時可為 nil
	GainAccountID        string          `json:"gain_account_id"`            // 處分利得科目 e.g. 4205
	LossAccountID        string          `json:"loss_account_id"`            // 處分損失科目 e.g. 5601
	Memo                 string          `json:"memo,omitempty"`
}

func (p AssetDisposedPayload) Validate() error {
	var errs []string
	if p.AssetUUID == "" {
		errs = append(errs, "asset_uuid is required")
	}
	if p.DisposalDate == "" {
		errs = append(errs, "disposal_date is required")
	}
	if p.Proceeds.LessThan(decimal.Zero) {
		errs = append(errs, "proceeds must be >= 0")
	}
	if p.Proceeds.GreaterThan(decimal.Zero) && (p.ProceedsLedgerUUID == nil || *p.ProceedsLedgerUUID == "") {
		errs = append(errs, "proceeds_ledger_uuid is required when proceeds > 0")
	}
	if p.GainAccountID == "" {
		errs = append(errs, "gain_account_id is required")
	}
	if p.LossAccountID == "" {
		errs = append(errs, "loss_account_id is required")
	}
	return joinErrors(errs)
}

// DepreciationAmount calculates this period's straight-line depreciation.
// Last period gets the remainder to avoid decimal drift.
func DepreciationAmount(cost, residualValue decimal.Decimal, usefulLifeMonths, depreciatedPeriods int64, totalDepreciated decimal.Decimal) decimal.Decimal {
	depreciableAmount := cost.Sub(residualValue)
	remaining := usefulLifeMonths - depreciatedPeriods
	if remaining <= 1 {
		return depreciableAmount.Sub(totalDepreciated)
	}
	return depreciableAmount.Div(decimal.NewFromInt(usefulLifeMonths)).Truncate(6)
}

// BuildAssetPurchasedTransaction assembles the journal entry payload for EventAssetPurchased.
// Called by the pipeline factory; the result is stored in AssetPurchasedState.Transaction.
func BuildAssetPurchasedTransaction(p AssetPurchasedPayload, category *projection.FixedAssetCategory, ledger *projection.LedgerAccount) (TransactionCreatedPayload, error) {
	var entries []TransactionEntryPayload
	switch p.PaymentType.Val() {
	case enums.AssetPaymentTypeCash:
		ledgerId := ledger.LedgerId
		entries = []TransactionEntryPayload{
			{AccountId: category.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero},
			{AccountId: ledger.AccountId, LedgerId: &ledgerId, Debit: decimal.Zero, Credit: p.Cost},
		}
	case enums.AssetPaymentTypeLease:
		entries = []TransactionEntryPayload{
			{AccountId: category.AssetAccountID, Debit: p.Cost, Credit: decimal.Zero},
			{AccountId: p.LiabilityAccountID, Debit: decimal.Zero, Credit: p.Cost},
		}
	default:
		return TransactionCreatedPayload{}, fmt.Errorf("invalid payment type: %s", p.PaymentType.Val())
	}
	return TransactionCreatedPayload{
		TransactionDate: p.PurchaseDate,
		Description:     fmt.Sprintf("固定資產購入 %s", p.Name),
		Currency:        "TWD",
		Entries:         entries,
	}, nil
}

// BuildAssetDepreciatedTransaction assembles the journal entry payload for EventAssetDepreciated.
// Called by the pipeline factory; the result is stored in AssetDepreciatedState.Transaction.
func BuildAssetDepreciatedTransaction(p AssetDepreciatedPayload, asset *projection.FixedAsset) TransactionCreatedPayload {
	deprAmount := DepreciationAmount(asset.Cost, asset.ResidualValue, asset.UsefulLifeMonths, asset.DepreciatedPeriods, asset.TotalDepreciated)
	return TransactionCreatedPayload{
		TransactionDate: p.PeriodDate + "-01",
		Description:     fmt.Sprintf("固定資產折舊 %s %s", asset.Name, p.PeriodDate),
		Currency:        "TWD",
		Entries: []TransactionEntryPayload{
			{AccountId: asset.DepreciationExpenseAccountID, Debit: deprAmount, Credit: decimal.Zero},
			{AccountId: asset.AccumDepreciationAccountID, Debit: decimal.Zero, Credit: deprAmount},
		},
	}
}

// BuildAssetDisposedTransaction assembles the journal entry payload for EventAssetDisposed.
// Called by the pipeline factory; the result is stored in AssetDisposedState.Transaction.
func BuildAssetDisposedTransaction(p AssetDisposedPayload, asset *projection.FixedAsset, proceedsLedger *projection.LedgerAccount) TransactionCreatedPayload {
	bookValue := asset.Cost.Sub(asset.TotalDepreciated)
	gainLoss := p.Proceeds.Sub(bookValue)
	entries := []TransactionEntryPayload{
		{AccountId: asset.AccumDepreciationAccountID, Debit: asset.TotalDepreciated, Credit: decimal.Zero},
		{AccountId: asset.AssetAccountID, Debit: decimal.Zero, Credit: asset.Cost},
	}
	if p.Proceeds.GreaterThan(decimal.Zero) && proceedsLedger != nil {
		ledgerId := proceedsLedger.LedgerId
		entries = append(entries, TransactionEntryPayload{
			AccountId: proceedsLedger.AccountId,
			LedgerId:  &ledgerId,
			Debit:     p.Proceeds,
			Credit:    decimal.Zero,
		})
	}
	if gainLoss.GreaterThan(decimal.Zero) {
		entries = append(entries, TransactionEntryPayload{
			AccountId: p.GainAccountID,
			Credit:    gainLoss,
			Debit:     decimal.Zero,
		})
	} else if gainLoss.LessThan(decimal.Zero) {
		entries = append(entries, TransactionEntryPayload{
			AccountId: p.LossAccountID,
			Debit:     gainLoss.Abs(),
			Credit:    decimal.Zero,
		})
	}
	return TransactionCreatedPayload{
		TransactionDate: p.DisposalDate,
		Description:     fmt.Sprintf("固定資產處分 %s", asset.Name),
		Currency:        "TWD",
		Entries:         entries,
	}
}
