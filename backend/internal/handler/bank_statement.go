package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/db/projection"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/csvparser"
	"akatengu/internal/pkg/xlsxparser"
	"akatengu/internal/repos/query"
	"akatengu/internal/repos/unit_of_work/event_store"
	"akatengu/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type bankStatementHandler struct {
	tmplSvc   services.BankStatementTemplateService
	svc       services.BankStatementService
	matchSvc  services.BankStatementMatchService
	reviewSvc services.BankStatementReviewService
	es        *services.EventStoreService
	l         *zap.Logger
}

// ─────────────────────────────────────────
// CSV Template handlers
// ─────────────────────────────────────────

func (h *bankStatementHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	method := "get bank csv templates"
	result, err := h.tmplSvc.GetAllTemplates(r.Context())
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankStatementHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "create bank csv template"
	ctx := r.Context()

	var p payload.BankCsvTemplateCreatedPayload
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
		AggregateType:   enums.AggregateBankCsvTemplate.Enum(),
		AggregateID:     newID.String(),
		ExpectedVersion: 0,
		EventType:       event_types.EventBankCsvTemplateCreated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankStatementHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "update bank csv template"
	ctx := r.Context()
	templateID := r.PathValue("template_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
		payload.BankCsvTemplateUpdatedPayload
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	req.TemplateUUID = templateID
	if err := req.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	b, err := json.Marshal(req.BankCsvTemplateUpdatedPayload)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateBankCsvTemplate.Enum(),
		AggregateID:     templateID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventBankCsvTemplateUpdated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankStatementHandler) DeactivateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "deactivate bank csv template"
	ctx := r.Context()
	templateID := r.PathValue("template_id")

	var req struct {
		ExpectedVersion int64 `json:"expected_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	p := payload.BankCsvTemplateDeactivatedPayload{TemplateUUID: templateID}
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
		AggregateType:   enums.AggregateBankCsvTemplate.Enum(),
		AggregateID:     templateID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventBankCsvTemplateDeactivated.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

// ─────────────────────────────────────────
// Import handlers
// ─────────────────────────────────────────

func (h *bankStatementHandler) GetImportsPaged(w http.ResponseWriter, r *http.Request) {
	method := "get bank statement imports paged"
	ctx := r.Context()

	ledgerID, _ := strconv.ParseInt(r.URL.Query().Get("ledger_id"), 10, 64)
	page, _ := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	pageSize, _ := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	imports, total, err := h.svc.GetImportsPaged(ctx, ledgerID, page, pageSize)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, map[string]any{"data": imports, "total": total, "page": page, "size": pageSize})
}

func (h *bankStatementHandler) GetImportResult(w http.ResponseWriter, r *http.Request) {
	method := "get bank statement import result"
	ctx := r.Context()
	importID, err := strconv.ParseInt(r.PathValue("import_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid import_id")
		return
	}

	imp, err := h.svc.GetImportByID(ctx, importID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if imp == nil {
		response.WriteError(w, r, http.StatusNotFound, "Not Found", "import not found")
		return
	}

	txns, err := h.svc.GetTxnsByImport(ctx, importID)
	if err != nil {
		h.l.Error(method+" txns fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, map[string]any{"import": imp, "transactions": txns})
}

func (h *bankStatementHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	method := "import bank statement csv"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
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

	var templateID *int64
	if raw := r.FormValue("template_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err == nil && v > 0 {
			templateID = &v
		}
	}

	var note *string
	if n := r.FormValue("note"); n != "" {
		note = &n
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "file is required")
		return
	}
	defer file.Close()

	filename := header.Filename

	// Load template for column mapping
	var csvTmpl *projection.BankCsvTemplate
	if templateID != nil {
		csvTmpl, err = h.tmplSvc.GetTemplateByID(ctx, *templateID)
		if err != nil {
			h.l.Error(method+" load template fail", zap.Error(err))
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}
		if csvTmpl == nil {
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "template not found")
			return
		}
	} else {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "template_id is required")
		return
	}

	// Build parser template from DB template
	parserTmpl := buildParserTemplate(csvTmpl)

	// Parse CSV
	rows, err := csvparser.Parse(file, parserTmpl)
	if err != nil {
		h.l.Error(method+" parse csv fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", fmt.Sprintf("csv parse error: %s", err.Error()))
		return
	}
	if len(rows) == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "no transactions found in CSV")
		return
	}

	// Load template ledgers for multi-ledger support
	templateLedgers, err := h.tmplSvc.GetTemplateLedgers(ctx, *templateID)
	if err != nil {
		h.l.Error(method+" load template ledgers fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
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

	// Convert rows to payload items
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
		txnItems = append(txnItems, item)
	}

	templateUUID := csvTmpl.TemplateUUID
	p := payload.BankStatementImportedPayload{
		LedgerID:      ledgerID,
		TemplateUUID:  &templateUUID,
		TemplateID:    templateID,
		StatementDate: statementDate,
		ImportSource:  "CSV",
		Filename:      &filename,
		Note:          note,
		Ledgers:       importLedgers,
		Transactions:  txnItems,
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

func (h *bankStatementHandler) ImportExcel(w http.ResponseWriter, r *http.Request) {
	method := "import bank statement excel"
	ctx := r.Context()

	if err := r.ParseMultipartForm(10 << 20); err != nil {
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

	var templateID *int64
	if raw := r.FormValue("template_id"); raw != "" {
		v, err := strconv.ParseInt(raw, 10, 64)
		if err == nil && v > 0 {
			templateID = &v
		}
	}
	if templateID == nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "template_id is required")
		return
	}

	var note *string
	if n := r.FormValue("note"); n != "" {
		note = &n
	}

	password := r.FormValue("password")

	file, header, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "file is required")
		return
	}
	defer file.Close()

	filename := header.Filename

	csvTmpl, err := h.tmplSvc.GetTemplateByID(ctx, *templateID)
	if err != nil {
		h.l.Error(method+" load template fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if csvTmpl == nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "template not found")
		return
	}

	parserTmpl := buildParserTemplate(csvTmpl)

	rows, err := xlsxparser.Parse(file, parserTmpl, password)
	if err != nil {
		h.l.Error(method+" parse excel fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", fmt.Sprintf("excel parse error: %s", err.Error()))
		return
	}
	if len(rows) == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "no transactions found in Excel file")
		return
	}

	excelTemplateLedgers, err := h.tmplSvc.GetTemplateLedgers(ctx, *templateID)
	if err != nil {
		h.l.Error(method+" load template ledgers fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	excelImportLedgers := make([]payload.ImportLedgerItem, 0, len(excelTemplateLedgers))
	for _, tl := range excelTemplateLedgers {
		ilUUID, err := uuid.NewV7()
		if err != nil {
			h.l.Error(method+" generate uuid fail", zap.Error(err))
			response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
			return
		}
		excelImportLedgers = append(excelImportLedgers, payload.ImportLedgerItem{
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
		txnItems = append(txnItems, payload.BankStatementTxnItem{
			BankTxnUUID: txnUUID.String(),
			TxnDate:     row.TxnDate,
			Description: row.Description,
			Debit:       row.Debit,
			Credit:      row.Credit,
			Balance:     row.Balance,
			ReferenceNo: row.ReferenceNo,
		})
	}

	excelTemplateUUID := csvTmpl.TemplateUUID
	p := payload.BankStatementImportedPayload{
		LedgerID:      ledgerID,
		TemplateUUID:  &excelTemplateUUID,
		TemplateID:    templateID,
		StatementDate: statementDate,
		ImportSource:  "XLSX",
		Filename:      &filename,
		Note:          note,
		Ledgers:       excelImportLedgers,
		Transactions:  txnItems,
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

func buildParserTemplate(t *projection.BankCsvTemplate) csvparser.Template {
	return csvparser.Template{
		SkipRows:          int(t.SkipRows),
		DateColumn:        int(t.DateColumn),
		DateFormat:        t.DateFormat,
		DescriptionColumn: int(t.DescriptionColumn),
		DebitColumn:       toIntPtr(t.DebitColumn),
		CreditColumn:      toIntPtr(t.CreditColumn),
		AmountColumn:      toIntPtr(t.AmountColumn),
		BalanceColumn:     toIntPtr(t.BalanceColumn),
		ReferenceColumn:   toIntPtr(t.ReferenceColumn),
	}
}

func toIntPtr(v *int64) *int {
	if v == nil {
		return nil
	}
	n := int(*v)
	return &n
}

// ─────────────────────────────────────────
// Matching handlers
// ─────────────────────────────────────────

func (h *bankStatementHandler) AutoMatch(w http.ResponseWriter, r *http.Request) {
	method := "auto match bank statement"
	ctx := r.Context()
	importID, err := strconv.ParseInt(r.PathValue("import_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid import_id")
		return
	}
	result, err := h.matchSvc.AutoMatch(ctx, importID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankStatementHandler) ReMatch(w http.ResponseWriter, r *http.Request) {
	method := "re-match bank statement"
	ctx := r.Context()
	importID, err := strconv.ParseInt(r.PathValue("import_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid import_id")
		return
	}
	result, err := h.matchSvc.ReMatch(ctx, importID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *bankStatementHandler) ManualMatch(w http.ResponseWriter, r *http.Request) {
	method := "manual match bank statement txn"
	ctx := r.Context()
	bankTxnID, err := strconv.ParseInt(r.PathValue("bank_txn_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid bank_txn_id")
		return
	}

	var req struct {
		EntryID int64 `json:"entry_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.EntryID == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "entry_id is required")
		return
	}

	if err := h.matchSvc.ManualMatch(ctx, bankTxnID, req.EntryID); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

// ─────────────────────────────────────────
// Review handlers
// ─────────────────────────────────────────

func (h *bankStatementHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	method := "get bank statement review"
	ctx := r.Context()
	importID, err := strconv.ParseInt(r.PathValue("import_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid import_id")
		return
	}
	items, err := h.reviewSvc.GetReviewItems(ctx, importID)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, items)
}

func (h *bankStatementHandler) ApproveTxn(w http.ResponseWriter, r *http.Request) {
	method := "approve bank statement txn"
	ctx := r.Context()
	bankTxnID, err := strconv.ParseInt(r.PathValue("bank_txn_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid bank_txn_id")
		return
	}

	var req services.ApproveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	req.BankTxnID = bankTxnID

	if req.AccountID == "" || req.CounterAccountID == "" || req.LedgerID == 0 {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "account_id, counter_account_id, and ledger_id are required")
		return
	}

	if err := h.reviewSvc.ApproveTxn(ctx, req); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

func (h *bankStatementHandler) IgnoreTxn(w http.ResponseWriter, r *http.Request) {
	method := "ignore bank statement txn"
	ctx := r.Context()
	bankTxnID, err := strconv.ParseInt(r.PathValue("bank_txn_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid bank_txn_id")
		return
	}

	if err := h.reviewSvc.IgnoreTxn(ctx, bankTxnID); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

// ─────────────────────────────────────────
// Complete handler
// ─────────────────────────────────────────

func (h *bankStatementHandler) CompleteImport(w http.ResponseWriter, r *http.Request) {
	method := "complete bank statement import"
	ctx := r.Context()

	var req struct {
		ImportUUID      string `json:"import_uuid"`
		ExpectedVersion int64  `json:"expected_version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if req.ImportUUID == "" {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "import_uuid is required")
		return
	}

	p := payload.BankStatementCompletedPayload{ImportUUID: req.ImportUUID}
	b, err := json.Marshal(p)
	if err != nil {
		h.l.Error(method+" marshal fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	result, err := h.es.Append(ctx, cmd.AppendCmd{
		AggregateType:   enums.AggregateBankStatement.Enum(),
		AggregateID:     req.ImportUUID,
		ExpectedVersion: req.ExpectedVersion,
		EventType:       event_types.EventBankStatementCompleted.Enum(),
		Payload:         b,
	})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func newBankStatementHandler(db *sqlx.DB, es *services.EventStoreService, l *zap.Logger) *bankStatementHandler {
	qr := query.NewQueryRepository(db)
	uow := event_store.NewUnitOfWork(db)
	return &bankStatementHandler{
		tmplSvc:   services.NewBankStatementTemplateService(qr.BankCsvTemplate),
		svc:       services.NewBankStatementService(qr.BankStatementImport),
		matchSvc:  services.NewBankStatementMatchService(qr, uow),
		reviewSvc: services.NewBankStatementReviewService(qr, uow, es),
		es:        es,
		l:         l,
	}
}
