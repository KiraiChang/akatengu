package cache_service

import (
	"akatengu/internal/model/db/projection"
	"akatengu/internal/pkg/cache"
	"akatengu/internal/repos/query"
	"context"
	"encoding/json"
)

type AccountCache struct {
	client *cache.Client
	repo   query.AccountRepo
}

func NewAccountCache(client *cache.Client, repo query.AccountRepo) *AccountCache {
	return &AccountCache{
		client: client,
		repo:   repo,
	}
}

func (a *AccountCache) Get(ctx context.Context) []projection.Account {
	bytes, err := a.client.Get(ctx, "all_accounts", func(ctx context.Context) (any, error) {
		return a.repo.GetAllAccounts(ctx)
	})
	if err != nil {
		return nil
	}
	var accounts []projection.Account
	err = json.Unmarshal(bytes, &accounts)
	if err != nil {
		return nil
	}
	return accounts
}

func (a *AccountCache) GetIdMap(ctx context.Context) map[string]projection.Account {
	bytes, err := a.client.Get(ctx, "all_account_id_map", func(ctx context.Context) (any, error) {
		accounts := a.Get(ctx)
		m := make(map[string]projection.Account, len(accounts))
		for _, a := range accounts {
			m[a.AccountId] = a
		}
		return m, nil
	})
	if err != nil {
		return nil
	}
	var accounts map[string]projection.Account
	err = json.Unmarshal(bytes, &accounts)
	if err != nil {
		return nil
	}
	return accounts
}
