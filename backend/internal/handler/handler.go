package handler

import (
	"akatengu/internal/bootstrap"
	"akatengu/internal/enums"
	"akatengu/internal/handler/middleware"
	"akatengu/internal/pkg/jwt"
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

	jwt := jwt.NewJWT(cfg.JWT)
	auth := newAuthHandler(db, jwt, logger)

	account := newAccountHandler(db, logger)
	aggerate := newAggerateHandler(db, logger)
	txn := newTransactionHandler(db, logger)
	period := newPeriodHandler(db, logger)
	investment := newInvestmentHandler(db, logger)
	sys := newSysHandler(db, logger)

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
	api.HandleFunc("GET /account/all", account.GetAllAccount)
	api.HandleFunc("GET /ledger/paged", account.GetLedgerPaged)
	api.HandleFunc("GET /ledger/all", account.GetAllLedger)
	api.HandleFunc("GET /ledger/all_balance", account.GetAllLedgerBalances)
	api.HandleFunc("GET /account/{parent_id}", account.GetChildrenAccount)

	api.HandleFunc("GET /period/{period_type}", period.GetPeriodPageByType)

	api.HandleFunc("GET /investment/paged", investment.GetInvestmentPaged)
	api.HandleFunc("GET /investment/lot/paged", investment.GetOpenLotsPaged)
	api.HandleFunc("GET /investment/position", investment.GetPosition)
	api.HandleFunc("GET /investment/lot_disposal/paged", investment.GetLotDisposalsPaged)

	api.HandleFunc("GET /txn/paged", txn.GetTransactionPaged)
	api.HandleFunc("GET /txn/{txn_id}", txn.GetEntries)

	api.HandleFunc("GET /aggerate/{aggerate_type}", aggerate.GetVersion)

	api.HandleFunc("GET /sys/account", sys.GetSysAccount)
	api.HandleFunc("POST /sys/account", sys.UpdateSysAccount)

	// ★ 全域的middleware

	return mux
}
