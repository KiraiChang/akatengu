package services

import (
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"akatengu/internal/model/request"
	pkgjwt "akatengu/internal/pkg/jwt"
	"akatengu/internal/repos/query"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type MerchantService interface {
	Create(ctx context.Context, userID int64, req request.CreateMerchant) (int64, error)
	Select(ctx context.Context, userID int64, userName string, req request.SelectMerchant) (string, error)
	ListByUser(ctx context.Context, userID int64) ([]db.Merchant, error)
}

type merchantService struct {
	merchant query.MerchantRepo
	jwt      pkgjwt.JwtService
}

func NewMerchantService(merchant query.MerchantRepo, jwt pkgjwt.JwtService) MerchantService {
	return &merchantService{merchant: merchant, jwt: jwt}
}

func (s *merchantService) Create(ctx context.Context, userID int64, req request.CreateMerchant) (int64, error) {
	merchantID, err := s.merchant.Create(ctx, req.Name, req.DisplayName, req.Currency)
	if err != nil {
		return 0, fmt.Errorf("create merchant: %w", err)
	}
	if err := s.merchant.AddUser(ctx, userID, merchantID, enums.MerchantRoleTypeOwner.Enum()); err != nil {
		return 0, fmt.Errorf("add owner: %w", err)
	}
	return merchantID, nil
}

func (s *merchantService) Select(ctx context.Context, userID int64, userName string, req request.SelectMerchant) (string, error) {
	role, err := s.merchant.GetUserRole(ctx, userID, req.MerchantID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("not a member of this merchant")
	}
	if err != nil {
		return "", fmt.Errorf("get user role: %w", err)
	}

	user := &db.User{UserId: userID, Username: userName}
	token, err := s.jwt.GenerateTokenWithMerchant(user, req.MerchantID, role.String())
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

func (s *merchantService) ListByUser(ctx context.Context, userID int64) ([]db.Merchant, error) {
	return s.merchant.GetUserMerchants(ctx, userID)
}
