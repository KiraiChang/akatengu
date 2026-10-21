package query

import (
	"akatengu/internal/database/sqlcdb"
	"context"

	"github.com/jmoiron/sqlx"
)

// sqlxQuerier is the part of *sqlx.DB that repos running sqlx queries use.
// *sqlx.Tx satisfies it too, so those repos can be bound to a transaction
// (NewQueryRepositoryTx).
type sqlxQuerier interface {
	Rebind(query string) string
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
}

type Repo struct {
	Account             AccountRepo
	RunningBalance      RunningBalanceRepo
	Event               EventRepo
	Investment          InvestmentRepo
	Report              ReportRepo
	Period              PeriodRepo
	Entry               EntryRepo
	Transaction         TransactionRepo
	Installment         InstallmentRepo
	Sys                 SysRepo
	User                UserRepo
	Merchant            MerchantRepo
	Config              ConfigRepo
	Prepaid             PrepaidQueryRepo
	FixedAsset          FixedAssetQueryRepo
	FixedAssetCategory  FixedAssetCategoryQueryRepo
	PrepaidCategory     PrepaidCategoryQueryRepo
	Template            TemplateRepo
	AccountAnalysis     AccountAnalysisRepo
	Dashboard           DashboardRepo
	Audit               AuditRepo
	BankCsvTemplate     BankCsvTemplateRepo
	BankPdfTemplate     BankPdfTemplateRepo
	BankStatementImport BankStatementImportRepo
	EntryCFCategory     EntryCFCategoryRepo
}

func NewQueryRepository(db *sqlx.DB) *Repo {
	return newQueryRepository(sqlcdb.New(db), db, db)
}

// newQueryRepository builds every repo on q and sx. db is used only by the
// template repo, which opens its own transaction to write; it is nil for a
// Repo bound to a transaction.
func newQueryRepository(q *sqlcdb.Queries, sx sqlxQuerier, db *sqlx.DB) *Repo {
	return &Repo{
		Account:             newAccountRepo(q, sx),
		RunningBalance:      newRunningBalanceRepo(q),
		Event:               newEventRepo(q),
		Investment:          newInvestmentRepo(q, sx),
		Report:              newReportRepo(q, sx),
		Period:              newPeriodRepo(q),
		Entry:               newEntryRepo(q),
		Transaction:         newTransactionRepo(q),
		Installment:         newInstallmentRepo(q),
		Sys:                 newSysRepo(q),
		User:                newUserRepo(q),
		Merchant:            newMerchantRepo(q),
		Config:              newConfigRepo(q, sx),
		Prepaid:             newPrepaidQueryRepo(q),
		FixedAsset:          newFixedAssetQueryRepo(q),
		FixedAssetCategory:  newFixedAssetCategoryQueryRepo(q),
		PrepaidCategory:     newPrepaidCategoryQueryRepo(q),
		Template:            newTemplateRepo(q, db),
		AccountAnalysis:     newAccountAnalysisRepo(q),
		Dashboard:           newDashboardRepo(q, sx),
		Audit:               newAuditRepo(q),
		BankCsvTemplate:     newBankCsvTemplateRepo(q),
		BankPdfTemplate:     newBankPdfTemplateRepo(q),
		BankStatementImport: newBankStatementImportRepo(q),
		EntryCFCategory:     newEntryCFCategoryRepo(q),
	}
}
