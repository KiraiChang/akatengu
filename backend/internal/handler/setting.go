package handler

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/request"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/services"
	"encoding/json"
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

func (h *settingHandler) UpdateLedgerAccountTypeConfig(w http.ResponseWriter, r *http.Request) {
	method := "update ledger account type config"
	ctx := r.Context()

	typeStr := r.PathValue("type")
	t, err := enums.ParseLedgerAccountType(typeStr)
	if err != nil {
		h.l.Error(method+" parse type fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid ledger account type: "+typeStr)
		return
	}

	var req request.UpdateLedgerAccountTypeConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.l.Error(method+" validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	userName := ctxkey.GetUserName(ctx)
	var updatedBy *string
	if userName != "" {
		updatedBy = &userName
	}
	if err := h.s.UpdateLedgerAccountTypeConfig(ctx, t, req.AccountID, updatedBy); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
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

func (h *settingHandler) UpdateAssetTypeAccountConfig(w http.ResponseWriter, r *http.Request) {
	method := "update asset type account config"
	ctx := r.Context()

	typeStr := r.PathValue("type")
	at, err := enums.ParseAssetType(typeStr)
	if err != nil {
		h.l.Error(method+" parse type fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid asset type: "+typeStr)
		return
	}

	var req request.UpdateAssetTypeAccountConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.l.Error(method+" validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	userName := ctxkey.GetUserName(ctx)
	var updatedBy *string
	if userName != "" {
		updatedBy = &userName
	}
	params := sqlcdb.UpsertAssetTypeAccountConfigParams{
		AssetType:               at,
		RealizedGainAccountID:   req.RealizedGainAccountID,
		RealizedLossAccountID:   req.RealizedLossAccountID,
		UnrealizedGainAccountID: req.UnrealizedGainAccountID,
		UnrealizedLossAccountID: req.UnrealizedLossAccountID,
		OciAccountID:            req.OCIAccountID,
		FeeAccountID:            req.FeeAccountID,
		TaxAccountID:            req.TaxAccountID,
		AccountID:               req.AccountID,
		UpdatedBy:               updatedBy,
	}
	if err := h.s.UpdateAssetTypeAccountConfig(ctx, params); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}
