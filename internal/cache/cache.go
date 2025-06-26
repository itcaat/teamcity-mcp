package cache

import (
	"sync"
	"time"

	"github.com/itcaat/teamcity-mcp/internal/config"
	"github.com/itcaat/teamcity-mcp/internal/metrics"
)

// Cache provides in-memory caching with TTL
type Cache struct {
	data map[string]*cacheItem
	ttl  time.Duration
	mu   sync.RWMutex
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

// New creates a new cache instance
func New(cfg config.CacheConfig) (*Cache, error) {
	ttl, err := time.ParseDuration(cfg.TTL)
	if err != nil {
		return nil, err
	}

	cache := &Cache{
		data: make(map[string]*cacheItem),
		ttl:  ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache, nil
}

// Get retrieves a cached value
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		metrics.RecordCacheMiss("unknown")
		return nil, false
	}

	if time.Now().After(item.expiration) {
		metrics.RecordCacheMiss("expired")
		return nil, false
	}

	metrics.RecordCacheHit("hit")
	return item.value, true
}

// Set stores a value in the cache
func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(c.ttl),
	}
}

// Delete removes a value from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
}

// Clear removes all cached values
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheItem)
}

// cleanup removes expired items periodically
func (c *Cache) cleanup() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiration) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
