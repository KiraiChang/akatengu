package handler

import (
	"akatengu/internal/enums"
	"akatengu/internal/enums/event_types"
	"akatengu/internal/handler/response"
	"akatengu/internal/model/payload"
	"akatengu/internal/model/request/cmd"
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

// configAggregateID is the aggregate ID of the per-merchant ACCOUNT_CONFIG and
// SYS_CONFIG aggregates, the same one BootstrapEventSeeds writes the seed
// events to (bootstrap/event_seed.go). R2 of the BFS migration plan redefines
// these aggregates (ADR-024).
const configAggregateID = "1"

type settingHandler struct {
	s        services.SettingService
	es       *services.EventStoreService
	versions query.AggerateRepo
	l        *zap.Logger
}

func newSettingHandler(db *sqlx.DB, es *services.EventStoreService, logger *zap.Logger) *settingHandler {
	return &settingHandler{
		s:        services.NewSettingService(db),
		es:       es,
		versions: query.NewAggerateRepo(db),
		l:        logger,
	}
}

// appendConfigEvent appends a config event to the merchant's config aggregate.
// The client sends no version, so the current aggregate version is read right
// before appending, as the frontend does with getAggregateVersion.
func appendConfigEvent(ctx context.Context, es *services.EventStoreService, versions query.AggerateRepo,
	aggType enums.AggregateType, eventType event_types.EventType, p any) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	version, err := versions.GetVersion(ctx, aggType, configAggregateID)
	if errors.Is(err, sql.ErrNoRows) {
		version = 0
	} else if err != nil {
		return err
	}
	_, err = es.Append(ctx, cmd.AppendCmd{
		AggregateType:   aggType,
		AggregateID:     configAggregateID,
		ExpectedVersion: version,
		EventType:       eventType,
		Payload:         b,
	})
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
		payload.LedgerAccountTypeConfigUpdatedPayload
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

	err = appendConfigEvent(ctx, h.es, h.versions, enums.AggregateAccountConfig.Enum(),
		event_types.EventLedgerAccountTypeConfigUpdated.Enum(), p)
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
		payload.AssetTypeAccountConfigUpdatedPayload
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

	err = appendConfigEvent(ctx, h.es, h.versions, enums.AggregateAccountConfig.Enum(),
		event_types.EventAssetTypeAccountConfigUpdated.Enum(), p)
	if err != nil {
		h.l.Error(method+" fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	response.OK(w, nil)
}
