package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type ReportHandler struct {
	service *services.ReportService
	logger  *zap.Logger
}

func NewReportHandler(db *sqlx.DB, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		service: services.NewReportService(query.NewReportRepo(db)),
		logger:  logger,
	}
}

func (rh *ReportHandler) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reportDate := r.URL.Query().Get("report_date")
	if _, err := time.Parse("2006-01-02", reportDate); err != nil {
		rh.logger.Error("get balance sheet invalid date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "report_date must be YYYY-MM-DD")
		return
	}
	result, err := rh.service.GetBalanceSheet(ctx, reportDate)
	if err != nil {
		rh.logger.Error("get balance sheet fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, result)
}

func (rh *ReportHandler) GetIncomeStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	begin := r.URL.Query().Get("begin_date")
	if _, err := time.Parse("2006-01-02", begin); err != nil {
		rh.logger.Error("get income statement invalid begin_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "begin_date must be YYYY-MM-DD")
		return
	}

	end := r.URL.Query().Get("end_date")
	if _, err := time.Parse("2006-01-02", end); err != nil {
		rh.logger.Error("get income statement invalid end_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "end_date must be YYYY-MM-DD")
		return
	}

	result, err := rh.service.GetIncomeStatement(ctx, begin, end)
	if err != nil {
		rh.logger.Error("get income statement fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, result)
}

func (rh *ReportHandler) GetCashFlowStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	begin := r.URL.Query().Get("begin_date")
	if _, err := time.Parse("2006-01-02", begin); err != nil {
		rh.logger.Error("get cash flow statement invalid begin_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "begin_date must be YYYY-MM-DD")
		return
	}
	end := r.URL.Query().Get("end_date")
	if _, err := time.Parse("2006-01-02", end); err != nil {
		rh.logger.Error("get cash flow statement invalid end_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "end_date must be YYYY-MM-DD")
		return
	}
	result, err := rh.service.GetCashFlowStatement(ctx, begin, end)
	if err != nil {
		rh.logger.Error("get cash flow statement fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	response.OK(w, result)
}

func (rh *ReportHandler) GetEquityStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	begin := r.URL.Query().Get("begin_date")
	if _, err := time.Parse("2006-01-02", begin); err != nil {
		rh.logger.Error("get equity statement invalid begin_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "begin_date must be YYYY-MM-DD")
		return
	}
	end := r.URL.Query().Get("end_date")
	if _, err := time.Parse("2006-01-02", end); err != nil {
		rh.logger.Error("get equity statement invalid end_date", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "end_date must be YYYY-MM-DD")
		return
	}
	result, err := rh.service.GetEquityStatement(ctx, begin, end)
	if err != nil {
		rh.logger.Error("get equity statement fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	response.OK(w, result)
}
