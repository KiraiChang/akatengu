package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/model/db"
	"akatengu/internal/model/request"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type sysHandler struct {
	s services.SysService
	l *zap.Logger
}

func (h sysHandler) GetSysAccount(w http.ResponseWriter, r *http.Request) {
	method := "get sys account"
	ctx := r.Context()
	result, err := h.s.GetSysAccount(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h sysHandler) UpdateSysAccount(w http.ResponseWriter, r *http.Request) {
	method := "update sys account"
	ctx := r.Context()
	var req request.SysAccount
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
	sys := db.SysAccount{
		SysCode:   req.SysCode,
		AccountId: req.AccountId,
	}
	err := h.s.UpdateSysAccount(ctx, sys)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	response.OK(w, sys)
}

func newSysHandler(db *sqlx.DB, logger *zap.Logger) *sysHandler {
	return &sysHandler{
		s: services.NewSysService(db),
		l: logger,
	}
}
