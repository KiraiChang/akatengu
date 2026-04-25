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

func newAccountHandler(db *sqlx.DB) *accountHandler {
	return &accountHandler{
		account: services.NewAccountService(query.NewAccountRepo(db)),
		logger:  zap.NewNop(),
	}
}

func (h *accountHandler) GetAccountPaged(w http.ResponseWriter, r *http.Request) {
	method := "get account paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	result, err := h.account.GetAccountPaged(ctx, req)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *accountHandler) GetChildrenAccount(w http.ResponseWriter, r *http.Request) {
	method := "get children account"
	ctx := r.Context()
	q := r.URL.Query()

	parentId := q.Get("parent_id")

	result, err := h.account.GetChildrenAccount(ctx, parentId)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}
