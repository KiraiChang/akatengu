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

type installmentHandler struct {
	s services.InstallmentService
	l *zap.Logger
}

func (h installmentHandler) GetInstallmentPaged(w http.ResponseWriter, r *http.Request) {
	method := "get installment paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	result, total, err := h.s.GetInstallmentPaged(ctx, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h installmentHandler) GetPaymentPaged(w http.ResponseWriter, r *http.Request) {
	method := "get payment paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	paymentId, _ := strconv.ParseInt(q.Get("payment_id"), 10, 64)

	result, total, err := h.s.GetPaymentPaged(ctx, req, paymentId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func newInstallmentHandler(db *sqlx.DB, l *zap.Logger) *installmentHandler {
	return &installmentHandler{
		s: services.NewInstallmentService(query.NewInstallmentRepo(db)),
		l: l,
	}
}
