package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"context"

	"github.com/jmoiron/sqlx"
)

type UserRepo interface {
	GetByName(ctx context.Context, userName string) (*db.User, error)
	Create(ctx context.Context, user *db.User) (*db.User, error)
}

type userRepo struct {
	q *sqlcdb.Queries
}

func NewUserRepo(sqlxDB *sqlx.DB) UserRepo {
	return &userRepo{q: sqlcdb.New(sqlxDB)}
}

func newUserRepo(q *sqlcdb.Queries) UserRepo {
	return &userRepo{q: q}
}

func (u *userRepo) GetByName(ctx context.Context, userName string) (*db.User, error) {
	row, err := u.q.GetUserByName(ctx, sqlcdb.GetUserByNameParams{
		Username: userName,
		Status:   enums.UserStatusTypeActive.Enum(),
	})
	if err != nil {
		return nil, err
	}
	return db.UserPtrFromUser(row), nil
}

func (u *userRepo) Create(ctx context.Context, user *db.User) (*db.User, error) {
	id, err := u.q.CreateUser(ctx, sqlcdb.CreateUserParams{
		Username: user.Username,
		Password: user.Password,
	})
	if err != nil {
		return nil, err
	}
	user.UserId = id
	return user, nil
}
