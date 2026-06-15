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

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type entryCFCategoryHandler struct {
	repo query.EntryCFCategoryRepo
	txnQ query.TransactionRepo
	es   *services.EventStoreService
	l    *zap.Logger
}

func newEntryCFCategoryHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *entryCFCategoryHandler {
	qr := query.NewQueryRepository(db)
	return &entryCFCategoryHandler{
		repo: qr.EntryCFCategory,
		txnQ: qr.Transaction,
		es:   es,
		l:    l,
	}
}

func (h *entryCFCategoryHandler) GetCFReview(w http.ResponseWriter, r *http.Request) {
	method := "get cf review"
	ctx := r.Context()

	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")

	rows, err := h.repo.ListTransactionCFReview(ctx, dateFrom, dateTo)
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
		ExpectedVersion int64                          `json:"expected_version"`
		Entries         []payload.TransactionCFEntryItem `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	// validate entries cf_category values
	for i, e := range req.Entries {
		if !e.CFCategory.In(enums.CashFlowCategoryOperating, enums.CashFlowCategoryInvesting, enums.CashFlowCategoryFinancing) {
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request",
				"entries["+string(rune('0'+i))+"].cf_category must be OPERATING, INVESTING, or FINANCING")
			return
		}
	}

	p := payload.TransactionCFCategoryUpdatedPayload{
		TxnUUID: txnUUID,
		Entries: req.Entries,
	}
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
		AggregateType:   enums.AggregateTransaction.Enum(),
		AggregateID:     txnUUID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventTransactionCFCategoryUpdated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}
