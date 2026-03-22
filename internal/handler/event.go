package handler

import (
	"akatengu/internal/model/request/cmd"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

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

func (h *EventHandler) Append(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req cmd.AppendCmd
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("decode fail", zap.Error(err))
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	result, err := h.service.Append(ctx, req)
	if err != nil {
		h.logger.Error("create event fail", zap.Error(err))
		http.Error(w, "create event fail", http.StatusInternalServerError)
		return
	}

	// 回傳 JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
