package services

import (
	"akatengu/internal/model/request"
	"akatengu/internal/repos/query"
	"context"
	"fmt"
)

type AuthService interface {
	Login(ctx context.Context, req request.Login) (string, error)
}

type authService struct {
	user query.UserRepo
	jwt  JwtService
}

func (a *authService) Login(ctx context.Context, req request.Login) (string, error) {
	// DB 驗證帳密（略）
	user, err := a.user.GetByName(ctx, req.Username)
	if err != nil {
		return "", err
	}
	isVaild, err := VerifyPassword(req.Password, user.Password)
	if err != nil {
		return "", err
	}
	if !isVaild {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := a.jwt.GenerateToken(user)
	if err != nil {
		return "", err
	}

	return token, nil
}

func NewAuthService(user query.UserRepo, jwt JwtService) AuthService {
	return &authService{
		user: user,
		jwt:  jwt,
	}
}
