package projection_repo

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db/projection"
	"context"

	"github.com/shopspring/decimal"
)

type TransactionRepo interface {
	InsertTxn(ctx context.Context, p projection.Transaction) (int64, error)
	UpdateTxnStatus(ctx context.Context, txn_id int64, status enums.TransactionStatus, version int64, updatedBy *string) error
	SysUpdateTxnStatus(ctx context.Context, txn_id int64, refTxnId *int64, status enums.TransactionStatus, updatedBy *string) error
	UpsertJournalEntries(ctx context.Context, entries []projection.Entry) error
	GetEntriesByTxnId(ctx context.Context, txnId int64, merchantID int64) ([]projection.Entry, error)
}

type InstallmentRepo interface {
	// installment
	InsertInstallment(ctx context.Context, p *projection.Installment) (int64, error)
	InsertInstallmentPayment(ctx context.Context, p *projection.InstallmentPayment) (int64, error)
	UpdateInstallmentTxn(ctx context.Context, inst_id int64, txn_id int64, updatedBy *string) error
	PaidInstallmentPayment(ctx context.Context, id int64, date string, updatedBy *string) error
	UpdateInstallmentStatus(ctx context.Context, id int64, status enums.InstallmentStatus, updatedBy *string) error
	UpdatePaymentTxn(ctx context.Context, paymentId int64, txnId int64, updatedBy *string) error
}

type AccountRepo interface {
	// account
	CreateAccount(ctx context.Context, p projection.Account) error
	UpdateAccount(ctx context.Context, p projection.Account) error
	UpsertAccount(ctx context.Context, p projection.Account) error
	CreateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error
	UpdateLedgerAccount(ctx context.Context, p projection.LedgerAccount) error
}

//// 重建用
//TruncateAll(ctx context.Context) error
//TruncateByAggregate(ctx context.Context, aggregateType enums.AggregateType) error

type InvestmentRepo interface {
	// investment
	CreateInvestment(ctx context.Context, p projection.Investment) error
	UpdateInvestment(ctx context.Context, p projection.Investment) error
	UpsertExchangeRate(ctx context.Context, rate projection.ExchangeRate) error
	UpsertPosition(ctx context.Context, position projection.InvestmentPosition) error
	InsertLot(ctx context.Context, lot projection.InvestmentLot) (int64, error)
	InsertLotDisposals(ctx context.Context, disposals projection.InvestmentLotDisposals) (int64, error)
	InsertMovement(ctx context.Context, movement projection.InvestmentMovement) (int64, error)
	UpdateMovement(ctx context.Context, id int64, txnId int64, updatedBy *string) error
	UpdateLot(ctx context.Context, id int64, txnId int64, updatedBy *string) error
	PositionSplit(ctx context.Context, id int64, ratio decimal.Decimal, updatedBy *string) error
	LotSplit(ctx context.Context, id int64, ratio decimal.Decimal, updatedBy *string) error
	UpdateInvestmentPositionSold(ctx context.Context, position projection.InvestmentPosition) error
	UpdateDisposalTxn(ctx context.Context, lotId int64, txnId int64) error
	UpdatePositionFairValue(ctx context.Context, investmentId int64, marketPriceTWD decimal.Decimal, updatedBy *string) error
	UpdateLotUnrealizedUnit(ctx context.Context, lotId int64, unrealizedUnitTWD decimal.Decimal, updatedBy *string) error
}

type PeriodCloseRepo interface {
	InsertPeriodClose(ctx context.Context, c projection.PeriodClosing) (*int64, error)
	UpdatePeriodCloseSnapshotByPeriod(ctx context.Context, id int64, snapshot string, closedAt *string, updatedBy *string) error
	ReopenPeriodClose(ctx context.Context, id int64, reason string, at string, updatedBy *string) error
	UpdatePeriodCloseTxnId(ctx context.Context, id int64, closingTxnId *int64, openingTxnId *int64, updatedBy *string) error
}

type AccountBalanceSnapshotRepo interface {
	BulkInsert(ctx context.Context, merchantID int64, closingId int64) error
	DeleteByClosingId(ctx context.Context, merchantID int64, closingId int64) error
}

type LedgerAccountBalanceSnapshotRepo interface {
	BulkInsert(ctx context.Context, merchantID int64, closingId int64) error
	DeleteByClosingId(ctx context.Context, merchantID int64, closingId int64) error
}

type AccountRunningBalanceRepo interface {
	Upsert(ctx context.Context, accountId string, merchantID int64, debit, credit decimal.Decimal) error
	GetAncestorIds(ctx context.Context, accountId string, merchantID int64) ([]string, error)
}

type LedgerRunningBalanceRepo interface {
	Upsert(ctx context.Context, ledgerId int64, merchantID int64, debit, credit decimal.Decimal) error
}

type PrepaidRepo interface {
	InsertPrepaid(ctx context.Context, p sqlcdb.InsertPrepaidParams) (int64, error)
	UpdatePrepaidTxn(ctx context.Context, id int64, merchantID int64, txnID int64) error
	UpdatePrepaidAmortization(ctx context.Context, id int64, merchantID int64, deltaAmount decimal.Decimal, status enums.PrepaidStatus, updatedBy *string) error
	UpdatePrepaidDisposed(ctx context.Context, id int64, merchantID int64, updatedBy *string) error
	InsertPrepaidAmortization(ctx context.Context, p sqlcdb.InsertPrepaidAmortizationParams) error
}

type FixedAssetRepo interface {
	InsertFixedAsset(ctx context.Context, p sqlcdb.InsertFixedAssetParams) (int64, error)
	UpdateFixedAssetTxn(ctx context.Context, id int64, merchantID int64, txnID int64) error
	UpdateFixedAssetDepreciation(ctx context.Context, id int64, merchantID int64, deltaAmount decimal.Decimal, updatedBy *string) error
	UpdateFixedAssetDisposed(ctx context.Context, id int64, merchantID int64, disposalDate string, updatedBy *string) error
	InsertFixedAssetDepreciation(ctx context.Context, p sqlcdb.InsertFixedAssetDepreciationParams) error
}

type FixedAssetCategoryRepo interface {
	InsertFixedAssetCategory(ctx context.Context, p sqlcdb.InsertFixedAssetCategoryParams) error
	UpdateFixedAssetCategory(ctx context.Context, p sqlcdb.UpdateFixedAssetCategoryParams) error
	SoftDeleteFixedAssetCategory(ctx context.Context, p sqlcdb.SoftDeleteFixedAssetCategoryParams) error
}

type PrepaidCategoryRepo interface {
	InsertPrepaidCategory(ctx context.Context, p sqlcdb.InsertPrepaidCategoryParams) error
	UpdatePrepaidCategory(ctx context.Context, p sqlcdb.UpdatePrepaidCategoryParams) error
	SoftDeletePrepaidCategory(ctx context.Context, p sqlcdb.SoftDeletePrepaidCategoryParams) error
}

type BankCsvTemplateRepo interface {
	InsertBankCsvTemplate(ctx context.Context, p sqlcdb.InsertBankCsvTemplateParams) error
	UpdateBankCsvTemplate(ctx context.Context, p sqlcdb.UpdateBankCsvTemplateParams) error
	DeactivateBankCsvTemplate(ctx context.Context, p sqlcdb.DeactivateBankCsvTemplateParams) error
}

type BankStatementImportRepo interface {
	InsertBankStatementImport(ctx context.Context, p sqlcdb.InsertBankStatementImportParams) (int64, error)
	UpdateBankStatementImportStatus(ctx context.Context, p sqlcdb.UpdateBankStatementImportStatusParams) error
	InsertBankStatementTxn(ctx context.Context, p sqlcdb.InsertBankStatementTxnParams) error
	UpdateBankTxnMatch(ctx context.Context, p sqlcdb.UpdateBankTxnMatchParams) error
	UpdateBankTxnStatus(ctx context.Context, p sqlcdb.UpdateBankTxnStatusParams) error
	UpdateBankTxnCreatedTxn(ctx context.Context, p sqlcdb.UpdateBankTxnCreatedTxnParams) error
	ResetNonConfirmedMatches(ctx context.Context, p sqlcdb.ResetNonConfirmedMatchesParams) error
}
