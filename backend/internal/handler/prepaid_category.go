package handler

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/handler/response"
	kevent "akatengu/internal/kernel/event"
	dbmodel "akatengu/internal/model/db"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"context"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type prepaidCategoryHandler struct {
	s services.PrepaidCategoryService
	w *command.Writer
	l *zap.Logger
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

	var p domainevent.PrepaidCategoryCreatedPayload
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

func (h *prepaidCategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	method := "update prepaid category"
	ctx := r.Context()
	categoryID := r.PathValue("category_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
		domainevent.PrepaidCategoryUpdatedPayload
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

	result, err := appendCategoryEvent(ctx, h.w, r, req.PrepaidCategoryUpdatedPayload, req.ExpectedVersion)
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

	p := domainevent.PrepaidCategoryDeletedPayload{CategoryUUID: categoryID}
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

func newPrepaidCategoryHandler(db *sqlx.DB, w *command.Writer, l *zap.Logger) *prepaidCategoryHandler {
	return &prepaidCategoryHandler{
		s: services.NewPrepaidCategoryService(query.NewQueryRepository(db).PrepaidCategory),
		w: w,
		l: l,
	}
}

// appendCategoryEvent writes a prepaid or fixed asset category event through
// BFS (BFS R2 section 4) and returns its event_store record, the same response
// as the legacy path. expectedVersion is the version the screen read; the
// registry ignores it for a creation.
func appendCategoryEvent(ctx context.Context, w *command.Writer, r *http.Request, p kevent.Payload, expectedVersion int64) (*dbmodel.EventStore, error) {
	evt, err := command.Build(ctx, p, expectedVersion, kevent.EventParams{Source: "api", Host: r.Host})
	if err != nil {
		return nil, err
	}
	return w.Append(ctx, evt)
}
