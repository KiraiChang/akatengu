package handler

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/handler/response"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type fixedAssetCategoryHandler struct {
	s services.FixedAssetCategoryService
	w *command.Writer
	l *zap.Logger
}

func (h *fixedAssetCategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	method := "get all fixed asset categories"
	ctx := r.Context()
	result, err := h.s.GetAllFixedAssetCategories(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *fixedAssetCategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	method := "create fixed asset category"
	ctx := r.Context()

	var p domainevent.FixedAssetCategoryCreatedPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	// The registry names the new category's aggregate after the creation event
	// (BFS R2 D6), so no version is sent.
	result, err := appendCategoryEvent(ctx, h.w, r, p, 0)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *fixedAssetCategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	method := "update fixed asset category"
	ctx := r.Context()
	categoryID := r.PathValue("category_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
		domainevent.FixedAssetCategoryUpdatedPayload
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	req.CategoryUUID = categoryID
	if err := req.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	result, err := appendCategoryEvent(ctx, h.w, r, req.FixedAssetCategoryUpdatedPayload, req.ExpectedVersion)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *fixedAssetCategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	method := "delete fixed asset category"
	ctx := r.Context()
	categoryID := r.PathValue("category_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	p := domainevent.FixedAssetCategoryDeletedPayload{CategoryUUID: categoryID}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	result, err := appendCategoryEvent(ctx, h.w, r, p, req.ExpectedVersion)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func newFixedAssetCategoryHandler(db *sqlx.DB, w *command.Writer, l *zap.Logger) *fixedAssetCategoryHandler {
	return &fixedAssetCategoryHandler{
		s: services.NewFixedAssetCategoryService(query.NewQueryRepository(db).FixedAssetCategory),
		w: w,
		l: l,
	}
}
