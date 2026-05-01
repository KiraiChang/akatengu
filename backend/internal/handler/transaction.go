package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/services"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type transactionHandler struct {
	s services.TransactionService
	l *zap.Logger
}

func (h transactionHandler) GetTransactionPaged(w http.ResponseWriter, r *http.Request) {
	method := "get ledger paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	req.SetDefaults()

	result, total, err := h.s.GetTransactionPaged(ctx, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h transactionHandler) GetEntries(w http.ResponseWriter, r *http.Request) {
	method := "get children account"
	ctx := r.Context()
	txnIdVal := r.PathValue("txn_id")
	if txnIdVal == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "txn_id is required")
		return
	}

	txnId, err := strconv.ParseInt(txnIdVal, 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "txn_id is invalid")
		return
	}
	result, err := h.s.GetEntries(ctx, txnId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func newTransactionHandler(db *sqlx.DB, logger *zap.Logger) *transactionHandler {
	return &transactionHandler{
		s: services.NewTransactionService(db),
		l: logger,
	}
}
