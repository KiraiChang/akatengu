package query

import (
	"akatengu/internal/database/sqlcdb"
	"akatengu/internal/enums"
	"akatengu/internal/model/db"
	"context"
)

type MerchantRepo interface {
	Create(ctx context.Context, name, displayName, currency string) (int64, error)
	Get(ctx context.Context, merchantID int64) (*db.Merchant, error)
	GetUserMerchants(ctx context.Context, userID int64) ([]db.Merchant, error)
	GetUserRole(ctx context.Context, userID, merchantID int64) (enums.MerchantRoleType, error)
	AddUser(ctx context.Context, userID, merchantID int64, role enums.MerchantRoleType) error
}

type merchantRepo struct {
	q *sqlcdb.Queries
}

func newMerchantRepo(q *sqlcdb.Queries) MerchantRepo {
	return &merchantRepo{q: q}
}

func (r *merchantRepo) Create(ctx context.Context, name, displayName, currency string) (int64, error) {
	return r.q.CreateMerchant(ctx, sqlcdb.CreateMerchantParams{
		Name:        name,
		DisplayName: displayName,
		Currency:    currency,
	})
}

func (r *merchantRepo) Get(ctx context.Context, merchantID int64) (*db.Merchant, error) {
	row, err := r.q.GetMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	return db.MerchantPtrFromMerchant(row), nil
}

func (r *merchantRepo) GetUserMerchants(ctx context.Context, userID int64) ([]db.Merchant, error) {
	rows, err := r.q.GetUserMerchants(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]db.Merchant, len(rows))
	for i, row := range rows {
		result[i] = db.MerchantFromGetUserMerchantsRow(row)
	}
	return result, nil
}

func (r *merchantRepo) GetUserRole(ctx context.Context, userID, merchantID int64) (enums.MerchantRoleType, error) {
	return r.q.GetUserMerchant(ctx, sqlcdb.GetUserMerchantParams{
		UserID:     userID,
		MerchantID: merchantID,
	})
}

func (r *merchantRepo) AddUser(ctx context.Context, userID, merchantID int64, role enums.MerchantRoleType) error {
	return r.q.AddUserMerchant(ctx, sqlcdb.AddUserMerchantParams{
		UserID:     userID,
		MerchantID: merchantID,
		Role:       role,
	})
}
