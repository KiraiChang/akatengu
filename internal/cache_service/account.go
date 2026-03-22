package cache_service

import (
	"akatengu/internal/model/db"
	"akatengu/internal/pkg/cache"
	"akatengu/internal/repos"
	"context"
	"encoding/json"
)

type AccountCache struct {
	client *cache.Client
	repo   *repos.AccountRepo
}

func NewAccountCache(client *cache.Client, repo *repos.AccountRepo) *AccountCache {
	return &AccountCache{
		client: client,
		repo:   repo,
	}
}

func (a *AccountCache) Get(ctx context.Context) []db.Account {
	bytes, err := a.client.Get(ctx, "all_accounts", func(ctx context.Context) (any, error) {
		return a.repo.GetAllAccounts(ctx)
	})
	if err != nil {
		return nil
	}
	var accounts []db.Account
	err = json.Unmarshal(bytes, &accounts)
	if err != nil {
		return nil
	}
	return accounts
}

func (a *AccountCache) GetIdMap(ctx context.Context) map[int64]db.Account {
	bytes, err := a.client.Get(ctx, "all_account_id_map", func(ctx context.Context) (any, error) {
		accounts := a.Get(ctx)
		m := make(map[int64]db.Account, len(accounts))
		for _, a := range accounts {
			m[a.Id] = a
		}
		return m, nil
	})
	if err != nil {
		return nil
	}
	var accounts map[int64]db.Account
	err = json.Unmarshal(bytes, &accounts)
	if err != nil {
		return nil
	}
	return accounts
}
