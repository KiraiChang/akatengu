package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/pdfparser"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type bankPdfTemplateHandler struct {
	svc services.BankPdfTemplateService
	es  *services.EventStoreService
	l   *zap.Logger
}

func (h *bankPdfTemplateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	method := "get bank pdf templates"
	result, err := h.svc.GetAllTemplates(r.Context())
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankPdfTemplateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "create bank pdf template"
	ctx := r.Context()

	var p payload.BankPdfTemplateCreatedPayload
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
	p.TemplateUUID = newID.String()

	for i := range p.Ledgers {
		if p.Ledgers[i].TplLedgerUUID == "" {
			lid, err := uuid.NewV7()
			if err != nil {
				h.l.Error(method+" generate ledger uuid fail", zap.Error(err))
				response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
				return
			}
			p.Ledgers[i].TplLedgerUUID = lid.String()
		}
	}

	b, err := json.Marshal(p)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateBankPdfTemplate.Enum(),
		AggregateID:     p.TemplateUUID,
		ExpectedVersion: 0,
		EventType:       event_types.EventBankPdfTemplateCreated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankPdfTemplateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "update bank pdf template"
	ctx := r.Context()
	templateUUID := r.PathValue("template_uuid")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
		payload.BankPdfTemplateUpdatedPayload
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	req.TemplateUUID = templateUUID

	for i := range req.Ledgers {
		if req.Ledgers[i].TplLedgerUUID == "" {
			lid, err := uuid.NewV7()
			if err != nil {
				h.l.Error(method+" generate ledger uuid fail", zap.Error(err))
				response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
				return
			}
			req.Ledgers[i].TplLedgerUUID = lid.String()
		}
	}

	if err := req.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	b, err := json.Marshal(req.BankPdfTemplateUpdatedPayload)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateBankPdfTemplate.Enum(),
		AggregateID:     templateUUID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventBankPdfTemplateUpdated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankPdfTemplateHandler) DeactivateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "deactivate bank pdf template"
	ctx := r.Context()
	templateUUID := r.PathValue("template_uuid")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	p := payload.BankPdfTemplateDeactivatedPayload{TemplateUUID: templateUUID}
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
		AggregateType:   enums.AggregateBankPdfTemplate.Enum(),
		AggregateID:     templateUUID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventBankPdfTemplateDeactivated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankPdfTemplateHandler) ImportPDF(w http.ResponseWriter, r *http.Request) {
	method := "import bank statement pdf"
	ctx := r.Context()

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "failed to parse multipart form")
		return
	}

	ledgerID, err := strconv.ParseInt(r.FormValue("ledger_id"), 10, 64)
	if err != nil || ledgerID == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "ledger_id is required")
		return
	}
	statementDate := r.FormValue("statement_date")
	if statementDate == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "statement_date is required")
		return
	}
	templateUUID := r.FormValue("template_uuid")
	if templateUUID == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "template_uuid is required")
		return
	}

	var note *string
	if n := r.FormValue("note"); n != "" {
		note = &n
	}
	var password *string
	if pw := r.FormValue("password"); pw != "" {
		password = &pw
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "file is required")
		return
	}
	defer file.Close()
	filename := header.Filename

	tmpl, err := h.svc.GetTemplateByUUID(ctx, templateUUID)
	if err != nil {
		h.l.Error(method+" load template fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if tmpl == nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "pdf template not found")
		return
	}
	if !tmpl.IsActive {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "pdf template is inactive")
		return
	}

	templateLedgers, err := h.svc.GetTemplateLedgers(ctx, tmpl.TemplateID)
	if err != nil {
		h.l.Error(method+" load template ledgers fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	hints := make([]pdfparser.LedgerHint, 0, len(templateLedgers))
	for _, tl := range templateLedgers {
		hints = append(hints, pdfparser.LedgerHint{
			LedgerUUID:  tl.LedgerUUID,
			AccountType: tl.AccountType,
		})
	}

	parser, err := pdfparser.New(tmpl.BankType)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", fmt.Sprintf("unsupported bank type: %s", tmpl.BankType))
		return
	}

	rows, err := parser.Parse(file, hints, pdfparser.ParseOptions{Password: password})
	if err != nil {
		h.l.Error(method+" parse pdf fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", fmt.Sprintf("pdf parse error: %s", err.Error()))
		return
	}
	if len(rows) == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "no transactions found in PDF")
		return
	}

	importLedgers := make([]payload.ImportLedgerItem, 0, len(templateLedgers))
	for _, tl := range templateLedgers {
		ilUUID, err := uuid.NewV7()
		if err != nil {
			h.l.Error(method+" generate uuid fail", zap.Error(err))
			response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}
		importLedgers = append(importLedgers, payload.ImportLedgerItem{
			ImportLedgerUUID: ilUUID.String(),
			LedgerUUID:       tl.LedgerUUID,
			AccountType:      tl.AccountType,
		})
	}

	txnItems := make([]payload.BankStatementTxnItem, 0, len(rows))
	for _, row := range rows {
		txnUUID, err := uuid.NewV7()
		if err != nil {
			h.l.Error(method+" generate uuid fail", zap.Error(err))
			response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}
		item := payload.BankStatementTxnItem{
			BankTxnUUID: txnUUID.String(),
			TxnDate:     row.TxnDate,
			Description: row.Description,
			Debit:       row.Debit,
			Credit:      row.Credit,
			Balance:     row.Balance,
			ReferenceNo: row.ReferenceNo,
		}
		if row.LedgerUUID != "" {
			item.LedgerUUID = &row.LedgerUUID
		}
		txnItems = append(txnItems, item)
	}

	bankType := tmpl.BankType
	p := payload.BankStatementImportedPayload{
		LedgerID:        ledgerID,
		PdfTemplateUUID: &templateUUID,
		BankType:        &bankType,
		StatementDate:   statementDate,
		ImportSource:    "PDF",
		Filename:        &filename,
		Note:            note,
		Ledgers:         importLedgers,
		Transactions:    txnItems,
	}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	importUUID, err := uuid.NewV7()
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
		AggregateType:   enums.AggregateBankStatement.Enum(),
		AggregateID:     importUUID.String(),
		ExpectedVersion: 0,
		EventType:       event_types.EventBankStatementImported.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func newBankPdfTemplateHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *bankPdfTemplateHandler {
	qr := query.NewQueryRepository(db)
	return &bankPdfTemplateHandler{
		svc: services.NewBankPdfTemplateService(qr.BankPdfTemplate),
		es:  es,
		l:   l,
	}
}
