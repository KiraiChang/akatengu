package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type dashboardHandler struct {
	s services.DashboardService
	l *zap.Logger
}

func newDashboardHandler(db *sqlx.DB, l *zap.Logger) *dashboardHandler {
	return &dashboardHandler{
		s: services.NewDashboardService(db),
		l: l,
	}
}

func (h *dashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	method := "get dashboard summary"
	ctx := r.Context()

	today := time.Now().Format("2006-01-02")
	summary, err := h.s.GetSummary(ctx, today)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, summary)
}

func (h *dashboardHandler) GetMonthlyTrend(w http.ResponseWriter, r *http.Request) {
	method := "get dashboard monthly trend"
	ctx := r.Context()
	q := r.URL.Query()

	monthsStr := q.Get("months")
	months := int64(12)
	if monthsStr != "" {
		if m, err := strconv.ParseInt(monthsStr, 10, 64); err == nil && m > 0 && m <= 60 {
			months = m
		}
	}

	now := time.Now()
	toMonth := now.Format("2006-01")
	fromMonth := now.AddDate(0, -int(months-1), 0).Format("2006-01")

	items, err := h.s.GetMonthlyTrend(ctx, fromMonth, toMonth)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, items)
}

func (h *dashboardHandler) GetLedgerBalances(w http.ResponseWriter, r *http.Request) {
	method := "get ledger balances"
	ctx := r.Context()

	balances, err := h.s.GetLedgerBalances(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, balances)
}
