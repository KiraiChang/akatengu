package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type auditHandler struct {
	s  services.AuditService
	es *services.EventStoreService
	l  *zap.Logger
}

func newAuditHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *auditHandler {
	return &auditHandler{
		s:  services.NewAuditService(db),
		es: es,
		l:  l,
	}
}

func (h *auditHandler) GetAggregateVersions(w http.ResponseWriter, r *http.Request) {
	method := "get audit aggregate versions"
	ctx := r.Context()

	result, err := h.s.GetAggregateVersions(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *auditHandler) GetEventStorePaged(w http.ResponseWriter, r *http.Request) {
	method := "get audit event store paged"
	ctx := r.Context()
	q := r.URL.Query()

	page, _ := strconv.ParseInt(q.Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(q.Get("page_size"), 10, 64)
	req := model.PaginationParams{Page: page, PageSize: pageSize}
	req.SetDefaults()

	result, total, err := h.s.GetEventStorePaged(ctx, req)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, model.PaginateWithTotal(result, req, total))
}

func (h *auditHandler) GetCheckpoints(w http.ResponseWriter, r *http.Request) {
	method := "get audit checkpoints"
	ctx := r.Context()

	result, err := h.s.GetCheckpoints(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *auditHandler) GetSnapshots(w http.ResponseWriter, r *http.Request) {
	method := "get audit snapshots"
	ctx := r.Context()

	result, err := h.s.GetSnapshots(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *auditHandler) GetExchangeRates(w http.ResponseWriter, r *http.Request) {
	method := "get exchange rates"
	ctx := r.Context()

	currency := r.URL.Query().Get("currency")
	result, err := h.s.GetExchangeRates(ctx, currency)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

type replayRequest struct {
	FromEventID   int64   `json:"from_event_id"`
	AggregateType *string `json:"aggregate_type"`
}

type replayResponse struct {
	ReplayedCount int `json:"replayed_count"`
}

func (h *auditHandler) Replay(w http.ResponseWriter, r *http.Request) {
	method := "replay projections"
	ctx := r.Context()

	var req replayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	var aggregateType *enums.AggregateType
	if req.AggregateType != nil {
		at, err := enums.ParseAggregateType(*req.AggregateType)
		if err != nil {
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid aggregate_type: "+*req.AggregateType)
			return
		}
		aggregateType = &at
	}

	events, err := h.es.Replay(ctx, req.FromEventID, aggregateType)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, replayResponse{ReplayedCount: len(events)})
}
