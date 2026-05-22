package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type prepaidHandler struct {
	s services.PrepaidService
	l *zap.Logger
}

func (h *prepaidHandler) GetAllPrepaids(w http.ResponseWriter, r *http.Request) {
	method := "get all prepaids"
	ctx := r.Context()

	result, err := h.s.GetAllPrepaids(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *prepaidHandler) GetActivePrepaids(w http.ResponseWriter, r *http.Request) {
	method := "get active prepaids"
	ctx := r.Context()

	result, err := h.s.GetActivePrepaids(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *prepaidHandler) GetPrepaidAmortizations(w http.ResponseWriter, r *http.Request) {
	method := "get prepaid amortizations"
	ctx := r.Context()

	prepaidID, _ := strconv.ParseInt(r.PathValue("prepaid_id"), 10, 64)

	result, err := h.s.GetPrepaidAmortizations(ctx, prepaidID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func newPrepaidHandler(db *sqlx.DB, l *zap.Logger) *prepaidHandler {
	return &prepaidHandler{
		s: services.NewPrepaidService(query.NewQueryRepository(db).Prepaid),
		l: l,
	}
}
