package cache

import (
	"sync"
	"time"
)

// InMemoryCache is a thread-safe, in-memory cache with TTL-based expiration.
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*memoryEntry
	hits    int
	misses  int
}

type memoryEntry struct {
	Result    *Result
	ExpiresAt time.Time
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		entries: make(map[string]*memoryEntry),
	}
}

// Get retrieves a cached result. Returns (result, true) on hit, (nil, false) on miss.
func (c *InMemoryCache) Get(url string) (*Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[url]
	if !ok {
		c.misses++
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(c.entries, url)
		c.misses++
		return nil, false
	}

	c.hits++
	return entry.Result, true
}

// Set stores a result in the cache with the given TTL.
func (c *InMemoryCache) Set(url string, result *Result, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[url] = &memoryEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete removes a specific URL from the cache.
func (c *InMemoryCache) Delete(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, url)
	return nil
}

// Clear removes all entries from the cache.
func (c *InMemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*memoryEntry)
	c.hits = 0
	c.misses = 0
	return nil
}

// Stats returns cache statistics.
func (c *InMemoryCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Hits:   c.hits,
		Misses: c.misses,
		Size:   len(c.entries),
	}
}
