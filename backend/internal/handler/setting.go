package handler

import (
	"akatengu/internal/application/command"
	domainevent "akatengu/internal/domain/event"
	"akatengu/internal/enums"
	"akatengu/internal/handler/response"
	kevent "akatengu/internal/kernel/event"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type settingHandler struct {
	s        services.SettingService
	w        *command.Writer
	versions query.AggerateRepo
	l        *zap.Logger
}

func newSettingHandler(db *sqlx.DB, w *command.Writer, logger *zap.Logger) *settingHandler {
	return &settingHandler{
		s:        services.NewSettingService(db),
		w:        w,
		versions: query.NewAggerateRepo(db),
		l:        logger,
	}
}

// appendConfigEvent writes a config event through BFS (BFS R2 section 3). The
// registry derives the merchant's config aggregate, aggType with
// domainevent.MerchantAggregateID, the same one BootstrapEventSeeds writes the
// seed events to. The client sends no version, so the current aggregate
// version is read right before writing, as the frontend does with
// getAggregateVersion.
func appendConfigEvent(ctx context.Context, w *command.Writer, versions query.AggerateRepo,
	aggType enums.AggregateType, p kevent.Payload, params kevent.EventParams) error {
	version, err := versions.GetVersion(ctx, aggType, domainevent.MerchantAggregateID)
	if errors.Is(err, sql.ErrNoRows) {
		version = 0
	} else if err != nil {
		return err
	}
	evt, err := command.Build(ctx, p, version, params)
	if err != nil {
		return err
	}
	_, err = w.Append(ctx, evt)
	return err
}

func (h *settingHandler) GetLedgerAccountTypeConfigs(w http.ResponseWriter, r *http.Request) {
	method := "get ledger account type configs"
	ctx := r.Context()
	result, err := h.s.GetLedgerAccountTypeConfigs(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *settingHandler) UpdateLedgerAccountTypeConfig(w http.ResponseWriter, r *http.Request) {
	method := "update ledger account type config"
	ctx := r.Context()

	typeStr := r.PathValue("type")
	t, err := enums.ParseLedgerAccountType(typeStr)
	if err != nil {
		h.l.Error(method+" parse type fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid ledger account type: "+typeStr)
		return
	}

	// The path decides the type. A "type" in the body is decoded into the
	// shallower Type field and ignored, so it is never parsed as an enum.
	var body struct {
		domainevent.LedgerAccountTypeConfigUpdatedPayload
		Type json.RawMessage `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	p := body.LedgerAccountTypeConfigUpdatedPayload
	p.Type = t
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	err = appendConfigEvent(ctx, h.w, h.versions, enums.AggregateAccountConfig.Enum(), p,
		kevent.EventParams{Source: "api", Host: r.Host})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}

func (h *settingHandler) GetAssetTypeAccountConfigs(w http.ResponseWriter, r *http.Request) {
	method := "get asset type account configs"
	ctx := r.Context()
	result, err := h.s.GetAssetTypeAccountConfigs(ctx)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, result)
}

func (h *settingHandler) UpdateAssetTypeAccountConfig(w http.ResponseWriter, r *http.Request) {
	method := "update asset type account config"
	ctx := r.Context()

	typeStr := r.PathValue("type")
	at, err := enums.ParseAssetType(typeStr)
	if err != nil {
		h.l.Error(method+" parse type fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", "invalid asset type: "+typeStr)
		return
	}

	// The path decides the asset type; see UpdateLedgerAccountTypeConfig.
	var body struct {
		domainevent.AssetTypeAccountConfigUpdatedPayload
		AssetType json.RawMessage `json:"asset_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.l.Error(method+" decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	p := body.AssetTypeAccountConfigUpdatedPayload
	p.AssetType = at
	if err := p.Validate(); err != nil {
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	err = appendConfigEvent(ctx, h.w, h.versions, enums.AggregateAccountConfig.Enum(), p,
		kevent.EventParams{Source: "api", Host: r.Host})
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}
