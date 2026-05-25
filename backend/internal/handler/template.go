package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/model/request"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type templateHandler struct {
	s services.TemplateService
	l *zap.Logger
}

func newTemplateHandler(db *sqlx.DB, l *zap.Logger) *templateHandler {
	return &templateHandler{
		s: services.NewTemplateService(db),
		l: l,
	}
}

func (h *templateHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	method := "get templates"
	ctx := r.Context()

	result, err := h.s.GetTemplates(ctx, r.URL.Query().Get("q"))
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *templateHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	method := "get template"
	ctx := r.Context()

	id, err := strconv.ParseInt(r.PathValue("template_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid template_id")
		return
	}
	result, err := h.s.GetTemplate(ctx, id)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *templateHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "create template"
	ctx := r.Context()

	var req request.SaveTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.l.Error(method+" validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	userName := ctxkey.GetUserName(ctx)
	var updatedBy *string
	if userName != "" {
		updatedBy = &userName
	}
	result, err := h.s.CreateTemplate(ctx, req.ToCreateCmd(updatedBy))
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *templateHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	method := "update template"
	ctx := r.Context()

	id, err := strconv.ParseInt(r.PathValue("template_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid template_id")
		return
	}

	var req request.SaveTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.l.Error(method+" validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	userName := ctxkey.GetUserName(ctx)
	var updatedBy *string
	if userName != "" {
		updatedBy = &userName
	}
	if err := h.s.UpdateTemplate(ctx, id, req.ToUpdateCmd(updatedBy)); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

func (h *templateHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	method := "delete template"
	ctx := r.Context()

	id, err := strconv.ParseInt(r.PathValue("template_id"), 10, 64)
	if err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid template_id")
		return
	}
	if err := h.s.DeleteTemplate(ctx, id); err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}
