package services

import (
	"akatengu/internal/model/db"
	"akatengu/internal/model/request"
	"akatengu/internal/repos/query"
	"context"
)

type UserService interface {
	CreateUser(ctx context.Context, req request.CreateUser) error
}

type userService struct {
	repo query.UserRepo
}

func (u *userService) CreateUser(ctx context.Context, req request.CreateUser) error {
	hash, err := HashPassword(req.Password)
	if err != nil {
		return err
	}
	user := &db.User{
		Username: req.Username,
		Password: hash,
	}
	user, err = u.repo.Create(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func NewUserService(repo query.UserRepo) UserService {
	return &userService{
		repo,
	}
}
