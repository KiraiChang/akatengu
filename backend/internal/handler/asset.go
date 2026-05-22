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

type fixedAssetHandler struct {
	s services.FixedAssetService
	l *zap.Logger
}

func (h *fixedAssetHandler) GetAllFixedAssets(w http.ResponseWriter, r *http.Request) {
	method := "get all fixed assets"
	ctx := r.Context()

	result, err := h.s.GetAllFixedAssets(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *fixedAssetHandler) GetActiveFixedAssets(w http.ResponseWriter, r *http.Request) {
	method := "get active fixed assets"
	ctx := r.Context()

	result, err := h.s.GetActiveFixedAssets(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h *fixedAssetHandler) GetFixedAssetDepreciations(w http.ResponseWriter, r *http.Request) {
	method := "get fixed asset depreciations"
	ctx := r.Context()

	assetID, _ := strconv.ParseInt(r.PathValue("asset_id"), 10, 64)

	result, err := h.s.GetFixedAssetDepreciations(ctx, assetID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func newFixedAssetHandler(db *sqlx.DB, l *zap.Logger) *fixedAssetHandler {
	return &fixedAssetHandler{
		s: services.NewFixedAssetService(query.NewQueryRepository(db).FixedAsset),
		l: l,
	}
}
