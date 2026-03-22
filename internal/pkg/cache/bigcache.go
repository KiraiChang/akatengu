package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/allegro/bigcache/v3"
)

type BigCacheLayer[T any] struct {
	cache *bigcache.BigCache
}

func newBigCacheLayer[T any](ctx context.Context, duration time.Duration) (*BigCacheLayer[T], error) {
	bc, err := bigcache.New(ctx, bigcache.DefaultConfig(duration))
	if err != nil {
		return nil, err
	}
	return &BigCacheLayer[T]{cache: bc}, nil
}

func (b *BigCacheLayer[T]) Get(ctx context.Context, key string) (T, bool, error) {
	var zero T

	buf, err := b.cache.Get(key)
	if err != nil {
		return zero, false, nil
	}

	var val T
	if err := json.Unmarshal(buf, &val); err != nil {
		return zero, false, err
	}

	return val, true, nil
}

func (b *BigCacheLayer[T]) Set(ctx context.Context, key string, val T) error {
	buf, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return b.cache.Set(key, buf)
}
