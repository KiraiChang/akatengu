package cache

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

type localCache struct {
	cache *ristretto.Cache[string, []byte]
	ttl   time.Duration
}

func newLocalCache(
	ttl time.Duration,
	maxItems int64,
) (*localCache, error) {

	c, err := ristretto.NewCache(&ristretto.Config[string, []byte]{
		NumCounters: maxItems * 10,
		MaxCost:     maxItems,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &localCache{cache: c, ttl: ttl}, nil
}

func (l *localCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	raw, ok := l.cache.Get(key)
	if !ok {
		return nil, false, nil
	}
	return raw, true, nil
}

func (l *localCache) Set(ctx context.Context, key string, val []byte) error {
	cost := int64(len(key) + binary.Size(val))
	l.cache.SetWithTTL(key, val, cost, l.ttl)
	return nil
}
