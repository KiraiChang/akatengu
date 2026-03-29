package handler

import (
	"akatengu/internal/model/enums"
	"akatengu/internal/repos"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/services/projection"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewMux(db *sqlx.DB, logger *zap.Logger) *http.ServeMux {
	// 4. DI（手動）
	//builder := cache.NewBuilder()
	//builder.WithLocalCache(time.Minute * 10)
	//client, err := builder.Build()
	//if err != nil {
	//	logger.Error("building client fail", zap.Error(err))
	//	panic(err)
	//}

	enums.InitEnums()

	uow := event_store.NewUnitOfWork(db)
	queryRepo := query.NewQueryRepository(db)
	projections := []projection.Projection{
		&projection.TransactionProjectionService{},
		&projection.AccountProjectionService{},
		&projection.InvestmentProjectionService{},
		&projection.PeriodProjectionService{},
	}
	eventService := services.NewEventStoreService(uow, queryRepo, projections)
	event := NewEventHandler(eventService, logger)

	reportRepo := repos.NewReportRepo(db)
	reportService := services.NewReportService(reportRepo)
	report := NewReportHandler(reportService, logger)

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	mux.HandleFunc("POST /event/append", event.Append)
	mux.HandleFunc("GET /report/balance_sheet", report.GetBalanceSheet)
	mux.HandleFunc("GET /report/income_statement", report.GetIncomeStatement)

	// ★ 全域的middleware

	return mux
}
