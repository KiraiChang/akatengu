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
	merchant := newMerchantHandler(db, jwt, logger)

	account := newAccountHandler(db, logger)
	aggerate := newAggerateHandler(db, logger)
	txn := newTransactionHandler(db, logger)
	period := newPeriodHandler(db, logger)
	investment := newInvestmentHandler(db, logger)
	sys := newSysHandler(db, logger)
	setting := newSettingHandler(db, logger)
	installment := newInstallmentHandler(db, logger)

	mux := http.NewServeMux()
	// SPA：所有其他請求
	mux.Handle("/", web.SPAHandler(web.FileSystem()))

	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK!"))
	}))

	mux.HandleFunc("POST /api/user/create", auth.CreateUser)
	mux.HandleFunc("POST /api/auth/login", auth.Login)

	// 需要 JWT 但不需商戶 context 的路由（商戶管理）
	api := middleware.NewGroup(mux, "/api", middleware.Jwt(jwt))
	api.HandleFunc("POST /merchant/create", merchant.Create)
	api.HandleFunc("GET /merchant/list", merchant.List)
	api.HandleFunc("POST /merchant/select", merchant.Select)

	// 需要 JWT + 商戶 context 的業務路由
	bizApi := api.SubGroup("", middleware.MerchantAuth(queryRepo.Merchant))
	bizApi.HandleFunc("POST /event/append", event.Append)

	bizApi.HandleFunc("GET /report/balance_sheet", report.GetBalanceSheet)
	bizApi.HandleFunc("GET /report/income_statement", report.GetIncomeStatement)
	bizApi.HandleFunc("GET /report/cash_flow_statement", report.GetCashFlowStatement)
	bizApi.HandleFunc("GET /report/cash_flow_statement_direct", report.GetDirectCashFlowStatement)
	bizApi.HandleFunc("GET /report/equity_statement", report.GetEquityStatement)

	bizApi.HandleFunc("GET /account/paged", account.GetAccountPaged)
	bizApi.HandleFunc("GET /account/all", account.GetAllAccount)
	bizApi.HandleFunc("GET /account/{parent_id}", account.GetChildrenAccount)
	bizApi.HandleFunc("GET /account/all_balance", account.GetAllAccountBalances)
	bizApi.HandleFunc("GET /ledger/paged", account.GetLedgerPaged)
	bizApi.HandleFunc("GET /ledger/all", account.GetAllLedger)
	bizApi.HandleFunc("GET /ledger/all_balance", account.GetAllLedgerBalances)

	bizApi.HandleFunc("GET /period/{period_type}", period.GetPeriodPageByType)

	bizApi.HandleFunc("GET /investment/paged", investment.GetInvestmentPaged)
	bizApi.HandleFunc("GET /investment/lot/paged", investment.GetOpenLotsPaged)
	bizApi.HandleFunc("GET /investment/position", investment.GetPosition)
	bizApi.HandleFunc("GET /investment/lot_disposal/paged", investment.GetLotDisposalsPaged)
	bizApi.HandleFunc("GET /investment/movement/paged", investment.GetMovementPaged)

	bizApi.HandleFunc("GET /installment/paged", installment.GetInstallmentPaged)
	bizApi.HandleFunc("GET /installment/payment", installment.GetPaymentPaged)

	bizApi.HandleFunc("GET /txn/paged", txn.GetTransactionPaged)
	bizApi.HandleFunc("GET /txn/{txn_id}", txn.GetEntries)

	bizApi.HandleFunc("GET /aggerate/{aggerate_type}", aggerate.GetVersion)

	bizApi.HandleFunc("GET /sys/account", sys.GetSysAccount)
	bizApi.HandleFunc("POST /sys/account", sys.UpdateSysAccount)

	bizApi.HandleFunc("GET /setting/ledger-account-type", setting.GetLedgerAccountTypeConfigs)
	bizApi.HandleFunc("PUT /setting/ledger-account-type/{type}", setting.UpdateLedgerAccountTypeConfig)
	bizApi.HandleFunc("GET /setting/asset-type", setting.GetAssetTypeAccountConfigs)
	bizApi.HandleFunc("PUT /setting/asset-type/{type}", setting.UpdateAssetTypeAccountConfig)

	return mux
}
