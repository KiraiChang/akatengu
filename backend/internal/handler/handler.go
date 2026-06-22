package handler

import (
	"akatengu/internal/bootstrap"
	"akatengu/internal/handler/middleware"
	"akatengu/internal/persistence/handle"
	"akatengu/internal/persistence/repos"
	"akatengu/internal/pkg/jwt"
	"akatengu/internal/process"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/runtime/engine"
	"akatengu/internal/runtime/mediator"
	"akatengu/internal/runtime/safety"
	"akatengu/internal/services"
	"akatengu/internal/web"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func NewMux(db *sqlx.DB, cfg bootstrap.Config, logger *zap.Logger) *http.ServeMux {
	// 1. Mediator
	m := mediator.NewMediator().WithQuery(db)

	process.Registry(m)

	// 2. event store
	store := repos.NewUnitOfWork(db)

	// 3. projector
	projector := handle.NewRegistry()

	// 4. executor
	executor := engine.NewExecutor(m)

	// 4. bfs
	bfs := engine.NewBFSEngine(executor).
		WithMiddleware([]safety.Middleware{safety.BackpressureScheduler(3), safety.DepthGuard(10), safety.CycleDetector()}...).
		WithProjector(projector).
		WithUnitOfWork(store)
	// 4. DI（手動）
	//builder := cache.NewBuilder()
	//builder.WithLocalCache(time.Minute * 10)
	//client, err := builder.Build()
	//if err != nil {
	//	logger.Error("building client fail", zap.Error(err))
	//	panic(err)
	//}

	uow := event_store.NewUnitOfWork(db)
	queryRepo := query.NewQueryRepository(db)
	eventService := services.NewEventStoreService(uow, queryRepo)
	event := NewEventHandler(eventService, logger)

	report := NewReportHandler(db, logger)

	jwt := jwt.NewJWT(cfg.JWT)
	auth := newAuthHandler(db, jwt, logger)
	merchant := newMerchantHandler(db, jwt, logger)

	account := newAccountHandler(db, logger)
	accountAnalysis := newAccountAnalysisHandler(db, logger)
	aggerate := newAggerateHandler(db, logger)
	txn := newTransactionHandler(db, logger)
	period := newPeriodHandler(db, logger)
	investment := newInvestmentHandler(db, logger)
	sys := newSysHandler(db, logger)
	setting := newSettingHandler(db, logger)
	installment := newInstallmentHandler(db, logger)
	prepaid := newPrepaidHandler(db, logger)
	prepaidCategory := newPrepaidCategoryHandler(db, eventService, logger)
	fixedAsset := newFixedAssetHandler(db, logger)
	fixedAssetCategory := newFixedAssetCategoryHandler(db, eventService, logger)
	template := newTemplateHandler(db, logger)
	dashboard := newDashboardHandler(db, logger)
	audit := newAuditHandler(db, eventService, logger)
	bankStatement := newBankStatementHandler(db, eventService, logger)
	bankPdfTemplate := newBankPdfTemplateHandler(db, eventService, logger)
	entryCFCategory := newEntryCFCategoryHandler(db, eventService, logger, bfs)

	mux := http.NewServeMux()
	// SPA：所有其他請求
	mux.Handle("/", web.SPAHandler(web.FileSystem()))

	mux.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK!"))
	}))

	mux.HandleFunc("POST /api/user/create", auth.CreateUser)
	mux.HandleFunc("POST /api/auth/login", auth.Login)

	// 需要 JWT 但不需商戶 context 的路由（商戶管理）
	api := middleware.NewGroup(mux, "/api", middleware.RequestIDMiddleware, middleware.Jwt(jwt))
	api.HandleFunc("POST /merchant/create", merchant.Create)
	api.HandleFunc("GET /merchant/list", merchant.List)
	api.HandleFunc("POST /merchant/select", merchant.Select)
	api.HandleFunc("PUT /merchant/{merchant_id}", merchant.Update)
	api.HandleFunc("DELETE /merchant/{merchant_id}", merchant.Deactivate)

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
	bizApi.HandleFunc("GET /account/all_balance", account.GetAllAccountBalances)
	bizApi.HandleFunc("GET /account/{account_id}/children-balance", accountAnalysis.GetChildrenBalance)
	bizApi.HandleFunc("GET /account/{account_id}/entries", accountAnalysis.GetEntries)
	bizApi.HandleFunc("GET /account/{account_id}/monthly-balance", accountAnalysis.GetMonthlyBalance)
	bizApi.HandleFunc("GET /account/{parent_id}", account.GetChildrenAccount)
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

	bizApi.HandleFunc("GET /prepaid/category/all", prepaidCategory.GetAllCategories)
	bizApi.HandleFunc("POST /prepaid/category", prepaidCategory.CreateCategory)
	bizApi.HandleFunc("PUT /prepaid/category/{category_id}", prepaidCategory.UpdateCategory)
	bizApi.HandleFunc("DELETE /prepaid/category/{category_id}", prepaidCategory.DeleteCategory)
	bizApi.HandleFunc("GET /prepaid/all", prepaid.GetAllPrepaids)
	bizApi.HandleFunc("GET /prepaid/active", prepaid.GetActivePrepaids)
	bizApi.HandleFunc("GET /prepaid/{prepaid_id}/amortizations", prepaid.GetPrepaidAmortizations)

	bizApi.HandleFunc("GET /fixed_asset/category/all", fixedAssetCategory.GetAllCategories)
	bizApi.HandleFunc("POST /fixed_asset/category", fixedAssetCategory.CreateCategory)
	bizApi.HandleFunc("PUT /fixed_asset/category/{category_id}", fixedAssetCategory.UpdateCategory)
	bizApi.HandleFunc("DELETE /fixed_asset/category/{category_id}", fixedAssetCategory.DeleteCategory)
	bizApi.HandleFunc("GET /fixed_asset/all", fixedAsset.GetAllFixedAssets)
	bizApi.HandleFunc("GET /fixed_asset/active", fixedAsset.GetActiveFixedAssets)
	bizApi.HandleFunc("GET /fixed_asset/{asset_id}/depreciations", fixedAsset.GetFixedAssetDepreciations)

	bizApi.HandleFunc("GET /txn/paged", txn.GetTransactionPaged)
	bizApi.HandleFunc("GET /txn/{txn_id}", txn.GetEntries)
	bizApi.HandleFunc("GET /cf-categories", entryCFCategory.GetCFReview)
	bizApi.HandleFunc("PUT /txn/{txn_uuid}/cf-category", entryCFCategory.UpdateCFCategory)

	bizApi.HandleFunc("GET /aggerate/{aggerate_type}", aggerate.GetVersion)

	bizApi.HandleFunc("GET /sys/account", sys.GetSysAccount)

	bizApi.HandleFunc("GET /setting/ledger-account-type", setting.GetLedgerAccountTypeConfigs)
	bizApi.HandleFunc("GET /setting/asset-type", setting.GetAssetTypeAccountConfigs)

	bizApi.HandleFunc("GET /template", template.GetTemplates)
	bizApi.HandleFunc("POST /template", template.CreateTemplate)
	bizApi.HandleFunc("GET /template/{template_id}", template.GetTemplate)
	bizApi.HandleFunc("PUT /template/{template_id}", template.UpdateTemplate)
	bizApi.HandleFunc("DELETE /template/{template_id}", template.DeleteTemplate)

	bizApi.HandleFunc("GET /dashboard/summary", dashboard.GetSummary)
	bizApi.HandleFunc("GET /dashboard/monthly-trend", dashboard.GetMonthlyTrend)
	bizApi.HandleFunc("GET /ledger/balances", dashboard.GetLedgerBalances)

	bizApi.HandleFunc("GET /audit/aggregate-version", audit.GetAggregateVersions)
	bizApi.HandleFunc("GET /audit/event", audit.GetEventStorePaged)
	bizApi.HandleFunc("GET /audit/checkpoint", audit.GetCheckpoints)
	bizApi.HandleFunc("GET /audit/snapshot", audit.GetSnapshots)
	bizApi.HandleFunc("GET /exchange-rate", audit.GetExchangeRates)
	bizApi.HandleFunc("POST /audit/replay", audit.Replay)
	bizApi.HandleFunc("POST /audit/event/export", event.Export)
	bizApi.HandleFunc("POST /audit/event/import", event.Import)

	bizApi.HandleFunc("GET /bank-statement/template", bankStatement.GetTemplates)
	bizApi.HandleFunc("POST /bank-statement/template", bankStatement.CreateTemplate)
	bizApi.HandleFunc("PUT /bank-statement/template/{template_id}", bankStatement.UpdateTemplate)
	bizApi.HandleFunc("DELETE /bank-statement/template/{template_id}", bankStatement.DeactivateTemplate)

	bizApi.HandleFunc("GET /bank-pdf-template", bankPdfTemplate.GetTemplates)
	bizApi.HandleFunc("POST /bank-pdf-template", bankPdfTemplate.CreateTemplate)
	bizApi.HandleFunc("GET /bank-pdf-template/{template_uuid}/ledgers", bankPdfTemplate.GetTemplateLedgers)
	bizApi.HandleFunc("PUT /bank-pdf-template/{template_uuid}", bankPdfTemplate.UpdateTemplate)
	bizApi.HandleFunc("DELETE /bank-pdf-template/{template_uuid}", bankPdfTemplate.DeactivateTemplate)

	bizApi.HandleFunc("POST /bank-statement/import", bankStatement.ImportCSV)
	bizApi.HandleFunc("POST /bank-statement/import/excel", bankStatement.ImportExcel)
	bizApi.HandleFunc("POST /bank-statement/import/pdf", bankPdfTemplate.ImportPDF)
	bizApi.HandleFunc("GET /bank-statement/paged", bankStatement.GetImportsPaged)
	bizApi.HandleFunc("GET /bank-statement/{import_id}/result", bankStatement.GetImportResult)
	bizApi.HandleFunc("POST /bank-statement/{import_id}/auto-match", bankStatement.AutoMatch)
	bizApi.HandleFunc("POST /bank-statement/{import_id}/re-match", bankStatement.ReMatch)
	bizApi.HandleFunc("PUT /bank-statement/{import_id}/txn/{bank_txn_id}/match", bankStatement.ManualMatch)
	bizApi.HandleFunc("GET /bank-statement/{import_id}/review", bankStatement.GetReview)
	bizApi.HandleFunc("POST /bank-statement/{import_id}/txn/{bank_txn_id}/approve", bankStatement.ApproveTxn)
	bizApi.HandleFunc("POST /bank-statement/{import_id}/txn/{bank_txn_id}/ignore", bankStatement.IgnoreTxn)
	bizApi.HandleFunc("POST /bank-statement/{import_id}/complete", bankStatement.CompleteImport)

	return mux
}
