package query

import "github.com/jmoiron/sqlx"

// Repo 跟query的合併為一個repos
type Repo struct {
	Event      EventRepo
	Investment InvestmentRepo
	Report     ReportRepo
	Period     PeriodRepo
	Entry      EntryRepo
}

func NewQueryRepository(db *sqlx.DB) *Repo {
	return &Repo{
		Event:      NewEventRepo(db),
		Investment: NewInvestmentRepo(db),
		Report:     NewReportRepo(db),
		Period:     NewPeriodRepo(db),
		Entry:      NewEntryRepo(db),
	}
}
