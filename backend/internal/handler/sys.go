package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/services"
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

func newSysHandler(db *sqlx.DB, logger *zap.Logger) *sysHandler {
	return &sysHandler{
		s: services.NewSysService(db),
		l: logger,
	}
}
