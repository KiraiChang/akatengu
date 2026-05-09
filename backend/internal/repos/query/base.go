package query

import (
	"akatengu/internal/database/sqlcdb"

	"github.com/jmoiron/sqlx"
)

type Repo struct {
	Account     AccountRepo
	Event       EventRepo
	Investment  InvestmentRepo
	Report      ReportRepo
	Period      PeriodRepo
	Entry       EntryRepo
	Transaction TransactionRepo
	Installment InstallmentRepo
	Sys         SysRepo
	User        UserRepo
}

func NewQueryRepository(db *sqlx.DB) *Repo {
	q := sqlcdb.New(db)
	return &Repo{
		Account:     newAccountRepo(q, db),
		Event:       newEventRepo(q),
		Investment:  newInvestmentRepo(q, db),
		Report:      newReportRepo(q),
		Period:      newPeriodRepo(q),
		Entry:       newEntryRepo(q),
		Transaction: newTransactionRepo(q),
		Installment: newInstallmentRepo(q),
		Sys:         newSysRepo(q),
		User:        newUserRepo(q),
	}
}
