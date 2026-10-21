package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/payload"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type sysHandler struct {
	s        services.SysService
	es       *services.EventStoreService
	versions query.AggerateRepo
	l        *zap.Logger
}

func (h sysHandler) GetSysAccount(w http.ResponseWriter, r *http.Request) {
	method := "get sys account"
	ctx := r.Context()
	result, err := h.s.GetSysAccount(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	response.OK(w, result)
}

func (h sysHandler) UpdateSysAccount(w http.ResponseWriter, r *http.Request) {
	method := "update sys account"
	ctx := r.Context()

	var p payload.SysAccountUpdatedPayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	err := appendConfigEvent(ctx, h.es, h.versions, enums.AggregateSysConfig.Enum(),
		event_types.EventSysAccountUpdated.Enum(), p)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

func newSysHandler(db *sqlx.DB, es *services.EventStoreService, logger *zap.Logger) *sysHandler {
	return &sysHandler{
		s:        services.NewSysService(db),
		es:       es,
		versions: query.NewAggerateRepo(db),
		l:        logger,
	}
}
