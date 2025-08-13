package cache

import (
	"context"
	"encoding/json"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/eko/gocache/store/go_cache/v4"
	"github.com/patrickmn/go-cache"
)

type CacheService interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, out any) error
	Delete(ctx context.Context, key string) error
}

type MemoryCache struct {
	cache *gocache.Cache[[]byte]
}

func NewMemoryCache(ttl, cleanupInterval time.Duration) CacheService {
	store := go_cache.NewGoCache(cache.New(ttl, cleanupInterval))

	return &MemoryCache{
		cache: gocache.New[[]byte](store),
	}
}
func (m *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return m.cache.Set(ctx, key, data, store.WithExpiration(ttl))
}

func (m *MemoryCache) Get(ctx context.Context, key string, out any) error {
	value, err := m.cache.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(value, out)
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	return m.cache.Delete(ctx, key)
}
