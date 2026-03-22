package cache

import (
	"context"
	"time"
)

type Builder struct {
	ctx           context.Context
	layers        []Layer
	flight        *FlightGroup
	policy        CachePolicy
	loaderTimeout time.Duration
}

func NewBuilder() *Builder {
	return &Builder{policy: WriteBack}
}

func (b *Builder) WithLocalCache(ttl time.Duration) *Builder {
	lc, _ := newLocalCache(ttl, 10_000)
	b.layers = append(b.layers, lc)
	return b
}

func (b *Builder) WithContext(ctx context.Context) *Builder {
	if ctx != nil {
		b.ctx = ctx
	}
	return b
}

func (b *Builder) WithSingleFlight() *Builder {
	b.flight = &FlightGroup{}
	return b
}

func (b *Builder) WithPolicy(p CachePolicy) *Builder {
	b.policy = p
	return b
}

func (b *Builder) WithLoaderTimeout(timeout time.Duration) *Builder {
	b.loaderTimeout = timeout
	return b
}

func (b *Builder) Build() (*Client, error) {
	return &Client{
		layers:        b.layers,
		flight:        b.flight,
		policy:        b.policy,
		loaderTimeout: b.loaderTimeout,
	}, nil
}
