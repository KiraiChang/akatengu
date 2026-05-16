package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/request"
	pkgjwt "akatengu/internal/pkg/jwt"
	"akatengu/internal/pkg/ctxkey"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type merchantHandler struct {
	svc    services.MerchantService
	logger *zap.Logger
}

func newMerchantHandler(db *sqlx.DB, jwt pkgjwt.JwtService, logger *zap.Logger) *merchantHandler {
	q := query.NewQueryRepository(db)
	svc := services.NewMerchantService(q.Merchant, jwt)
	return &merchantHandler{svc: svc, logger: logger}
}

func (h *merchantHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := ctx.Value(ctxkey.UserClaims).(*pkgjwt.Claims)

	var req request.CreateMerchant
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("create merchant decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("create merchant validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	merchantID, err := h.svc.Create(ctx, claims.UserID, req)
	if err != nil {
		h.logger.Error("create merchant fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, map[string]int64{"merchant_id": merchantID})
}

func (h *merchantHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := ctx.Value(ctxkey.UserClaims).(*pkgjwt.Claims)

	rows, err := h.svc.ListByUser(ctx, claims.UserID)
	if err != nil {
		h.logger.Error("list merchant fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, rows)
}

func (h *merchantHandler) Select(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims := ctx.Value(ctxkey.UserClaims).(*pkgjwt.Claims)

	var req request.SelectMerchant
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("select merchant decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		h.logger.Error("select merchant validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	token, err := h.svc.Select(ctx, claims.UserID, claims.UserName, req)
	if err != nil {
		h.logger.Error("select merchant fail", zap.Error(err))
		response.WriteError(w, r, http.StatusForbidden, "Forbidden", err.Error())
		return
	}

	response.OK(w, model.AuthResp{Token: token})
}
