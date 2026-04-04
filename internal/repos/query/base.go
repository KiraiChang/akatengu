package query

import "github.com/jmoiron/sqlx"

// Repo 跟query的合併為一個repos
type Repo struct {
	Account     AccountRepo
	Event       EventRepo
	Investment  InvestmentRepo
	Report      ReportRepo
	Period      PeriodRepo
	Entry       EntryRepo
	Transaction TransactionRepo
}

func NewQueryRepository(db *sqlx.DB) *Repo {
	return &Repo{
		Account:     NewAccountRepo(db),
		Event:       NewEventRepo(db),
		Investment:  NewInvestmentRepo(db),
		Report:      NewReportRepo(db),
		Period:      NewPeriodRepo(db),
		Entry:       NewEntryRepo(db),
		Transaction: NewTransactionRepo(db),
	}
}
