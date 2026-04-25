package cache

import (
	"context"
	"encoding/json"
	"time"
)

type Client struct {
	layers        []Layer
	flight        *FlightGroup
	policy        CachePolicy
	loaderTimeout time.Duration
}

func (c *Client) Get(
	ctx context.Context,
	key string,
	loader func(ctx context.Context) (any, error),
) ([]byte, error) {
	// ReadOnly or WriteBack 都先讀 cache
	for _, layer := range c.layers {
		if val, ok, err := layer.Get(ctx, key); err != nil {
			return nil, err
		} else if ok {
			return val, nil
		}
	}

	// ReadOnly：cache miss 就直接 loader，不寫回
	if c.policy == ReadOnly {
		return c.loadBinary(ctx, loader)
	}

	// WriteBack：singleflight + write-back
	load := func() ([]byte, error) {
		// 再次確認一次
		for _, layer := range c.layers {
			if val, ok, _ := layer.Get(ctx, key); ok {
				return val, nil
			}
		}

		// 3️⃣ 建立 loader timeout ctxkey（關鍵）
		lctx := ctx
		cancel := func() {}
		if c.loaderTimeout > 0 {
			lctx, cancel = context.WithTimeout(ctx, c.loaderTimeout)
		}
		defer cancel()

		bytes, err := c.loadBinary(lctx, loader)
		if err != nil {
			return nil, err
		}

		if bytes != nil {
			for _, layer := range c.layers {
				_ = layer.Set(ctx, key, bytes)
			}
		}
		return bytes, nil
	}

	if c.flight != nil {
		return c.flight.Do(key, load)
	}
	return load()
}

func (c *Client) loadBinary(ctx context.Context,
	loader func(ctx context.Context) (any, error),
) ([]byte, error) {
	result, err := loader(ctx)
	if err != nil {
		return nil, err
	}
	bytes, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}
