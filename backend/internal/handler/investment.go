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

type investmentHandler struct {
	s services.InvestmentService
	l *zap.Logger
}

func newInvestmentHandler(db *sqlx.DB, l *zap.Logger) *investmentHandler {
	return &investmentHandler{
		s: services.NewInvestmentService(query.NewInvestmentRepo(db)),
		l: l,
	}
}

func (h *investmentHandler) GetInvestmentPaged(w http.ResponseWriter, r *http.Request) {
	method := "get investment paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	result, total, err := h.s.GetInvestmentPaged(ctx, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *investmentHandler) GetOpenLotsPaged(w http.ResponseWriter, r *http.Request) {
	method := "get lot paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	txnId, _ := strconv.ParseInt(q.Get("txn_id"), 10, 64)

	result, total, err := h.s.GetOpenLotsPaged(ctx, req, txnId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *investmentHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	method := "get position"
	ctx := r.Context()
	q := r.URL.Query()

	investmentId, _ := strconv.ParseInt(q.Get("investment_id"), 10, 64)

	result, err := h.s.GetPosition(ctx, investmentId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *investmentHandler) GetLotDisposalsPaged(w http.ResponseWriter, r *http.Request) {
	method := "get lot disposal paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	lotId, _ := strconv.ParseInt(q.Get("lot_id"), 10, 64)

	result, total, err := h.s.GetLotDisposalsPaged(ctx, req, lotId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *investmentHandler) GetMovementPaged(w http.ResponseWriter, r *http.Request) {
	method := "get movement paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	investmentId, _ := strconv.ParseInt(q.Get("investment_id"), 10, 64)

	result, total, err := h.s.GetMovementPaged(ctx, req, investmentId)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}
