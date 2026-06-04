package handler

import (
	"akatengu/internal/handler/response"
	dbprojection "akatengu/internal/model/db/projection"
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/pkg/eventcrypto"
	"akatengu/internal/services"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type EventHandler struct {
	service *services.EventStoreService
	logger  *zap.Logger
}

func NewEventHandler(service *services.EventStoreService, logger *zap.Logger) *EventHandler {
	return &EventHandler{
		service: service,
		logger:  logger,
	}
}

func (h *EventHandler) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req struct {
		Password string `json:"password"`
	}
	// body 為選填；若為空 body 或解析失敗則視為無密碼匯出
	_ = json.NewDecoder(r.Body).Decode(&req)

	file, err := h.service.Export(ctx, req.Password)
	if err != nil {
		h.logger.Error("export events fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	filename := fmt.Sprintf("events-%s.json", time.Now().UTC().Format("20060102"))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(file); err != nil {
		h.logger.Error("encode export fail", zap.Error(err))
	}
}

func (h *EventHandler) Import(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "failed to parse multipart form")
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "missing file field")
		return
	}
	defer f.Close()

	var file dbprojection.EventExportFile
	if err := json.NewDecoder(f).Decode(&file); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid export file format")
		return
	}

	password := r.FormValue("password")
	count, err := h.service.Import(ctx, file, password)
	if err != nil {
		if errors.Is(err, services.ErrUnsupportedExportVersion) || errors.Is(err, services.ErrPasswordRequired) {
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
			return
		}
		if errors.Is(err, eventcrypto.ErrWrongPassword) {
			response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "wrong password or corrupted data")
			return
		}
		h.logger.Error("import events fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, map[string]int{"imported": count})
}

func (h *EventHandler) Append(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req cmd.AppendCmd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	result, err := h.service.Append(ctx, req)
	if err != nil {
		h.logger.Error("create event fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, result)
}