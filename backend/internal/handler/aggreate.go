package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type aggerateHandler struct {
	aggerate services.AggerateService
	logger   *zap.Logger
}

func newAggerateHandler(db *sqlx.DB, logger *zap.Logger) *aggerateHandler {
	return &aggerateHandler{
		aggerate: services.NewAggerateService(query.NewAggerateRepo(db)),
		logger:   logger,
	}
}

func (h *aggerateHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	method := "get version"
	ctx := r.Context()
	aggerateTypeVal := r.PathValue("aggerate_type")
	if aggerateTypeVal == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "aggerate_type is required")
		return
	}
	aggerateType, err := enums.ParseAggregateType(aggerateTypeVal)
	if err != nil {
		h.logger.Error(method+" fail, aggerate_type:"+aggerateTypeVal, zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid aggregate type")
		return
	}
	q := r.URL.Query()
	aggerateId := q.Get("aggerate_id")

	result, err := h.aggerate.GetVersion(ctx, aggerateType, aggerateId)
	if err != nil {
		h.logger.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}
