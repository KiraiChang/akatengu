package handler

import (
	"akatengu/internal/handler/response"
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