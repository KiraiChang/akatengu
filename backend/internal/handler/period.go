package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/services"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type periodHandler struct {
	s services.PeriodService
	l *zap.Logger
}

func (h periodHandler) GetPeriodPageByType(w http.ResponseWriter, r *http.Request) {
	method := "get period page  by type"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{
		Page:     int64(page),
		PageSize: int64(pageSize),
	}
	req.SetDefaults()

	periodTypeVal := r.PathValue("period_type")
	if periodTypeVal == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "period_type is required")
		return
	}
	periodType, err := enums.ParsePeriodType(periodTypeVal)
	if err != nil {
		h.l.Error(method+" fail, period_type:"+periodTypeVal, zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid period_type")
		return
	}

	result, total, err := h.s.GetPeriodPagedByType(ctx, periodType, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func newPeriodHandler(db *sqlx.DB, l *zap.Logger) *periodHandler {
	return &periodHandler{
		s: services.NewPeriodService(db),
		l: l,
	}
}
