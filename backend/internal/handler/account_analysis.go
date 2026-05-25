package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/services"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type accountAnalysisHandler struct {
	s services.AccountAnalysisService
	l *zap.Logger
}

func newAccountAnalysisHandler(db *sqlx.DB, l *zap.Logger) *accountAnalysisHandler {
	return &accountAnalysisHandler{
		s: services.NewAccountAnalysisService(db),
		l: l,
	}
}

func (h *accountAnalysisHandler) GetChildrenBalance(w http.ResponseWriter, r *http.Request) {
	method := "get account children balance"
	ctx := r.Context()

	accountID := r.PathValue("account_id")
	if accountID == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "missing account_id")
		return
	}

	children, err := h.s.GetAccountDirectChildrenWithBalance(ctx, accountID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, projection.AccountChildrenBalanceResult{
		AccountID: accountID,
		Name:      "",
		Children:  children,
	})
}

func (h *accountAnalysisHandler) GetEntries(w http.ResponseWriter, r *http.Request) {
	method := "get account entries"
	ctx := r.Context()
	q := r.URL.Query()

	accountID := r.PathValue("account_id")
	if accountID == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "missing account_id")
		return
	}

	fromDate := q.Get("from")
	toDate := q.Get("to")
	if fromDate == "" || toDate == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "from and to are required")
		return
	}

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
	req.SetDefaults()

	entries, total, err := h.s.GetAccountJournalEntriesPaged(ctx, accountID, fromDate, toDate, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(entries, req, total))
}

func (h *accountAnalysisHandler) GetMonthlyBalance(w http.ResponseWriter, r *http.Request) {
	method := "get account monthly balance"
	ctx := r.Context()
	q := r.URL.Query()

	accountID := r.PathValue("account_id")
	if accountID == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "missing account_id")
		return
	}

	fromMonth := q.Get("from")
	toMonth := q.Get("to")
	if fromMonth == "" || toMonth == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "from and to are required (format: YYYY-MM)")
		return
	}

	balances, err := h.s.GetAccountMonthlyBalances(ctx, accountID, fromMonth, toMonth)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, balances)
}
