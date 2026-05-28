package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/services"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type settingHandler struct {
	s services.SettingService
	l *zap.Logger
}

func newSettingHandler(db *sqlx.DB, logger *zap.Logger) *settingHandler {
	return &settingHandler{
		s: services.NewSettingService(db),
		l: logger,
	}
}

func (h *settingHandler) GetLedgerAccountTypeConfigs(w http.ResponseWriter, r *http.Request) {
	method := "get ledger account type configs"
	ctx := r.Context()
	result, err := h.s.GetLedgerAccountTypeConfigs(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *settingHandler) GetAssetTypeAccountConfigs(w http.ResponseWriter, r *http.Request) {
	method := "get asset type account configs"
	ctx := r.Context()
	result, err := h.s.GetAssetTypeAccountConfigs(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

