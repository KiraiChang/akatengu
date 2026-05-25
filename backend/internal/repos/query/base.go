package query

import (
	"akatengu/internal/database/sqlcdb"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	Account        AccountRepo
	RunningBalance RunningBalanceRepo
	Event          EventRepo
	Investment     InvestmentRepo
	Report         ReportRepo
	Period         PeriodRepo
	Entry          EntryRepo
	Transaction    TransactionRepo
	Installment    InstallmentRepo
	Sys            SysRepo
	User           UserRepo
	Merchant       MerchantRepo
	Config         ConfigRepo
	Prepaid        PrepaidQueryRepo
	FixedAsset     FixedAssetQueryRepo
	Template       TemplateRepo
}

func NewQueryRepository(db *sqlx.DB) *Repo {
	q := sqlcdb.New(db)
	return &Repo{
		Account:        newAccountRepo(q, db),
		RunningBalance: newRunningBalanceRepo(q),
		Event:          newEventRepo(q),
		Investment:     newInvestmentRepo(q, db),
		Report:         newReportRepo(q, db),
		Period:         newPeriodRepo(q),
		Entry:          newEntryRepo(q),
		Transaction:    newTransactionRepo(q),
		Installment:    newInstallmentRepo(q),
		Sys:            newSysRepo(q),
		User:           newUserRepo(q),
		Merchant:       newMerchantRepo(q),
		Config:         newConfigRepo(q, db),
		Prepaid:        newPrepaidQueryRepo(q),
		FixedAsset:     newFixedAssetQueryRepo(q),
		Template:       newTemplateRepo(q, db),
	}
}
