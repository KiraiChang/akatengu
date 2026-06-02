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

type prepaidCategoryHandler struct {
	s  services.PrepaidCategoryService
	es *services.EventStoreService
	l  *zap.Logger
}

func (h *prepaidCategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	method := "get all prepaid categories"
	ctx := r.Context()
	result, err := h.s.GetAllPrepaidCategories(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *prepaidCategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	method := "create prepaid category"
	ctx := r.Context()

	var p payload.PrepaidCategoryCreatedPayload
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
		AggregateType:   enums.AggregatePrepaidCategory.Enum(),
		AggregateID:     newID.String(),
		ExpectedVersion: 0,
		EventType:       event_types.EventPrepaidCategoryCreated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *prepaidCategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	method := "update prepaid category"
	ctx := r.Context()
	categoryID := r.PathValue("category_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
		payload.PrepaidCategoryUpdatedPayload
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

	b, err := json.Marshal(req.PrepaidCategoryUpdatedPayload)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregatePrepaidCategory.Enum(),
		AggregateID:     categoryID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventPrepaidCategoryUpdated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *prepaidCategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	method := "delete prepaid category"
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

	p := payload.PrepaidCategoryDeletedPayload{CategoryUUID: categoryID}
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
		AggregateType:   enums.AggregatePrepaidCategory.Enum(),
		AggregateID:     categoryID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventPrepaidCategoryDeleted.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func newPrepaidCategoryHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *prepaidCategoryHandler {
	return &prepaidCategoryHandler{
		s:  services.NewPrepaidCategoryService(query.NewQueryRepository(db).PrepaidCategory),
		es: es,
		l:  l,
	}
}
