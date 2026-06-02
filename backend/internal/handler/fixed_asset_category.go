package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type fixedAssetCategoryHandler struct {
	s  services.FixedAssetCategoryService
	es *services.EventStoreService
	l  *zap.Logger
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

	var p payload.FixedAssetCategoryCreatedPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	newID, err := uuid.NewV7()
	if err != nil {
		h.l.Error(method+" generate uuid fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
		AggregateID:     newID.String(),
		ExpectedVersion: 0,
		EventType:       event_types.EventFixedAssetCategoryCreated.Enum(),
		Payload:         b,
	})
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
		payload.FixedAssetCategoryUpdatedPayload
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

	b, err := json.Marshal(req.FixedAssetCategoryUpdatedPayload)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
		AggregateID:     categoryID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventFixedAssetCategoryUpdated.Enum(),
		Payload:         b,
	})
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

	p := payload.FixedAssetCategoryDeletedPayload{CategoryUUID: categoryID}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	b, err := json.Marshal(p)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateFixedAssetCategory.Enum(),
		AggregateID:     categoryID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventFixedAssetCategoryDeleted.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func newFixedAssetCategoryHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *fixedAssetCategoryHandler {
	return &fixedAssetCategoryHandler{
		s:  services.NewFixedAssetCategoryService(query.NewQueryRepository(db).FixedAssetCategory),
		es: es,
		l:  l,
	}
}
