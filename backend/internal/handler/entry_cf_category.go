package handler

import (
	event_payload "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	"akatengu/internal/kernel/event"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/runtime/engine"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type entryCFCategoryHandler struct {
	service services.EntryCFCategoryService
	txnQ    query.TransactionRepo
	es      *services.EventStoreService
	l       *zap.Logger
	bfs     *engine.BFSEngine
}

func newEntryCFCategoryHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger, engine *engine.BFSEngine) *entryCFCategoryHandler {
	qr := query.NewQueryRepository(db)
	return &entryCFCategoryHandler{
		service: services.NewEntryCFCategoryService(db),
		txnQ:    qr.Transaction,
		es:      es,
		l:       l,
		bfs:     engine,
	}
}

func (h *entryCFCategoryHandler) GetCFReview(w http.ResponseWriter, r *http.Request) {
	method := "get cf review"
	ctx := r.Context()

	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	rows, err := h.service.ListTransactionCFReview(ctx, dateFrom, dateTo)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}
	response.OK(w, rows)
}

func (h *entryCFCategoryHandler) UpdateCFCategory(w http.ResponseWriter, r *http.Request) {
	method := "update cf category"
	ctx := r.Context()
	txnUUID := r.PathValue("txn_uuid")

	var req struct {
		ExpectedVersion int64                                  `json:"expected_version"`
		Entries         []event_payload.TransactionCFEntryItem `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	evt := event.NewFullEvent[event_payload.TransactionCFCategoryUpdatedPayload](
		event.NewAggregateID(enums.AggregateTransaction.Enum()),
		event_payload.TransactionCFCategoryUpdatedPayload{
			TxnUUID: txnUUID,
			Entries: req.Entries,
		},
		event.EventParams{
			RequestID: ctxkey.GetRequestID(ctx),
			Source:    "api",
			Priority:  0,
			Host:      r.Host,
		})

	result, err := h.bfs.Run(ctx, evt)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
	}

	//// validate entries cf_category values
	//for i, e := range req.Entries {
	//	if !e.CFCategory.In(enums.CashFlowCategoryOperating, enums.CashFlowCategoryInvesting, enums.CashFlowCategoryFinancing) {
	//		response.WriteError(w, r, http.StatusBadRequest, "Bad Request",
	//			"entries["+string(rune('0'+i))+"].cf_category must be OPERATING, INVESTING, or FINANCING")
	//		return
	//	}
	//}
	//
	//p := payload.TransactionCFCategoryUpdatedPayload{
	//	TxnUUID: txnUUID,
	//	Entries: req.Entries,
	//}
	//if err := p.Validate(); err != nil {
	//	response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
	//	return
	//}
	//
	//b, err := json.Marshal(p)
	//if err != nil {
	//	h.l.Error(method+" marshal fail", zap.Error(err))
	//	response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
	//	return
	//}
	//
	//result, err := h.es.Append(ctx, cmd.AppendCmd{
	//	AggregateType:   enums.AggregateTransaction.Enum(),
	//	AggregateID:     txnUUID,
	//	ExpectedVersion: req.ExpectedVersion,
	//	EventType:       event_types.EventTransactionCFCategoryUpdated.Enum(),
	//	Payload:         b,
	//})
	//if err != nil {
	//	h.l.Error(method+" fail", zap.Error(err))
	//	response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
	//	return
	//}
	response.OK(w, result)
}
