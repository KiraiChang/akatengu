package handler

import (
	"akatengu/internal/services"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ReportHandler struct {
	service *services.ReportService
	logger  *zap.Logger
}

func NewReportHandler(service *services.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		logger:  logger,
	}
}

func (rh *ReportHandler) GetBalanceSheet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := "get balance sheet"
	reportDate := r.URL.Query().Get("report_date")
	_, err := time.Parse("2006-03-01", reportDate)
	if err != nil {
		rh.logger.Error(method+" fail", zap.Error(err))
		http.Error(w, method+" fail", http.StatusInternalServerError)
	}
	result, err := rh.service.GetBalanceSheet(ctx, reportDate)
	if err != nil {
		rh.logger.Error(method+" fail", zap.Error(err))
		http.Error(w, method+" fail", http.StatusInternalServerError)
		return
	}

	// 回傳 JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (rh *ReportHandler) GetIncomeStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	method := "get balance sheet"
	begin := r.URL.Query().Get("begin_date")
	_, err := time.Parse("2006-03-01", begin)
	if err != nil {
		rh.logger.Error(method+" fail", zap.Error(err))
		http.Error(w, method+" fail", http.StatusInternalServerError)
	}

	end := r.URL.Query().Get("end_date")
	_, err = time.Parse("2006-03-01", end)
	if err != nil {
		rh.logger.Error(method+" fail", zap.Error(err))
		http.Error(w, method+" fail", http.StatusInternalServerError)
	}
	result, err := rh.service.GetIncomeStatement(ctx, begin, end)
	if err != nil {
		rh.logger.Error(method+" fail", zap.Error(err))
		http.Error(w, method+" fail", http.StatusInternalServerError)
		return
	}

	// 回傳 JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
