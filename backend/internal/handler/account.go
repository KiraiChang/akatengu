package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type accountHandler struct {
	account services.AccountService
	logger  *zap.Logger
}

func newAccountHandler(db *sqlx.DB, logger *zap.Logger) *accountHandler {
	return &accountHandler{
		account: services.NewAccountService(query.NewAccountRepo(db), query.NewRunningBalanceRepo(db)),
		logger:  logger,
	}
}

func (h *accountHandler) GetAccountPaged(w http.ResponseWriter, r *http.Request) {
	method := "get account paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	result, total, err := h.account.GetAccountPaged(ctx, req)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *accountHandler) GetChildrenAccount(w http.ResponseWriter, r *http.Request) {
	method := "get children account"
	ctx := r.Context()
	parent_id := r.PathValue("parent_id")
	if parent_id == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "parent_id is required")
		return
	}
	result, err := h.account.GetChildrenAccount(ctx, parent_id)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *accountHandler) GetLedgerPaged(w http.ResponseWriter, r *http.Request) {
	method := "get ledger paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	result, total, err := h.account.GetLedgerPaged(ctx, req)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *accountHandler) GetAllAccount(w http.ResponseWriter, r *http.Request) {
	method := "get all account"
	ctx := r.Context()

	result, err := h.account.GetAllAccount(ctx)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *accountHandler) GetAllLedger(w http.ResponseWriter, r *http.Request) {
	method := "get all ledger"
	ctx := r.Context()

	result, err := h.account.GetAllLedgers(ctx)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *accountHandler) GetAllLedgerBalances(w http.ResponseWriter, r *http.Request) {
	method := "get all ledger balance"
	ctx := r.Context()

	result, err := h.account.GetAllLedgerBalances(ctx)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *accountHandler) GetAllAccountBalances(w http.ResponseWriter, r *http.Request) {
	method := "get all account balance"
	ctx := r.Context()

	result, err := h.account.GetAllAccountBalances(ctx)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}
