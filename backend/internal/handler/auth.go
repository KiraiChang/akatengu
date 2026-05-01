package handler

import (
	"akatengu/internal/handler/response"
	"akatengu/internal/handler/response/model"
	"akatengu/internal/model/request"
	"akatengu/internal/pkg/jwt"
	"akatengu/internal/repos/query"
	"akatengu/internal/services"
	"encoding/json"
	"net/http"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type authHandler struct {
	user   services.UserService
	auth   services.AuthService
	logger *zap.Logger
}

func newAuthHandler(db *sqlx.DB, jwt jwt.JwtService, logger *zap.Logger) *authHandler {
	userRepo := query.NewUserRepo(db)
	user := services.NewUserService(userRepo)
	auth := services.NewAuthService(userRepo, jwt)
	return &authHandler{user, auth, logger}
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req request.Login
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("login decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("login validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	result, err := h.auth.Login(ctx, req)
	if err != nil {
		h.logger.Error("login fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	response.OK(w, model.AuthResp{
		Token: result,
	})
}

func (h *authHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req request.CreateUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("create user decode fail", zap.Error(err))
		response.WriteError(w, r, http.StatusBadRequest, "Bad Request", err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Error("create user validate fail", zap.Error(err))
		response.WriteError(w, r, http.StatusUnprocessableEntity, "Unprocessable Entity", err.Error())
		return
	}

	if err := h.user.CreateUser(ctx, req); err != nil {
		h.logger.Error("create user fail", zap.Error(err))
		response.WriteError(w, r, http.StatusInternalServerError, "Internal Server Error", err.Error())
		return
	}

	response.OK(w, "OK")
}
