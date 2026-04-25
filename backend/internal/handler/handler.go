package handler

import (
	"akatengu/internal/bootstrap"
	"akatengu/internal/enums"
	"akatengu/internal/handler/middleware"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"akatengu/internal/web"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewMux(db *sqlx.DB, cfg bootstrap.Config, logger *zap.Logger) *http.ServeMux {
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
	eventService := services.NewEventStoreService(uow, queryRepo)
	event := NewEventHandler(eventService, logger)

	report := NewReportHandler(db, logger)

	jwt := services.NewJWT(cfg.JWT)
	auth := newAuthHandler(db, jwt, logger)

	account := newAccountHandler(db)

	mux := http.NewServeMux()
	// SPA：所有其他請求
	mux.Handle("/", web.SPAHandler(web.FileSystem()))

	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	mux.HandleFunc("POST /api/user/create", auth.CreateUser)
	mux.HandleFunc("POST /api/auth/login", auth.Login)

	api := middleware.NewGroup(mux, "/api", middleware.Jwt(jwt))
	api.HandleFunc("POST /event/append", event.Append)

	api.HandleFunc("GET /report/balance_sheet", report.GetBalanceSheet)
	api.HandleFunc("GET /report/income_statement", report.GetIncomeStatement)

	api.HandleFunc("GET /account/paged", account.GetAccountPaged)

	// ★ 全域的middleware

	return mux
}
