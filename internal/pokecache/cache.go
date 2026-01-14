package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	mu    sync.Mutex
	items map[string]cacheEntry
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		items: make(map[string]cacheEntry),
	}
	go cache.reapLoop(interval)

	return &cache
}

func (c *Cache) Add(key string, val []byte) {
	// this a write operation we have to lock it
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}

}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.items[key]
	if !ok {
		return nil, false
	}

	return item.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		c.mu.Lock()
		cutoff := time.Now().Add(-interval)
		for key, item := range c.items {
			if item.createdAt.Before(cutoff) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}
