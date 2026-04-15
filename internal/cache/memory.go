package cache

import (
	"sync"
	"time"
)

const (
	DefaultMaxSize = 1000
)

var DefaultMaxEntries = 10000

// InMemoryCache is a thread-safe, in-memory cache with TTL-based expiration
// and optional LRU eviction when MaxEntries is reached.
type InMemoryCache struct {
	mu          sync.RWMutex
	entries     map[string]*memoryEntry
	accessOrder []string
	maxEntries  int
	hits        int
	misses      int
}

type memoryEntry struct {
	Result    *Result
	ExpiresAt time.Time
}

// NewInMemoryCache creates a new in-memory cache with default limits.
// Default: 10000 max entries.
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		entries:     make(map[string]*memoryEntry),
		accessOrder: make([]string, 0, DefaultMaxEntries),
		maxEntries:  DefaultMaxEntries,
	}
}

// NewInMemoryCacheWithLimit creates a new in-memory cache with custom max entries.
// When maxEntries is reached, the least recently used entry is evicted.
func NewInMemoryCacheWithLimit(maxEntries int) *InMemoryCache {
	return &InMemoryCache{
		entries:     make(map[string]*memoryEntry),
		accessOrder: make([]string, 0, maxEntries),
		maxEntries:  maxEntries,
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
		c.removeFromAccessOrder(url)
		c.misses++
		return nil, false
	}

	c.hits++
	c.moveToEnd(url)
	return entry.Result, true
}

// Set stores a result in the cache with the given TTL.
func (c *InMemoryCache) Set(url string, result *Result, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[url]; ok {
		existing.Result = result
		existing.ExpiresAt = time.Now().Add(ttl)
		c.moveToEnd(url)
		return
	}

	if len(c.entries) >= c.maxEntries {
		c.evictLRU()
	}

	c.entries[url] = &memoryEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
	}
	c.accessOrder = append(c.accessOrder, url)
}

// Delete removes a specific URL from the cache.
func (c *InMemoryCache) Delete(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, url)
	c.removeFromAccessOrder(url)
	return nil
}

// Clear removes all entries from the cache.
func (c *InMemoryCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*memoryEntry)
	c.accessOrder = make([]string, 0, c.maxEntries)
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

func (c *InMemoryCache) moveToEnd(url string) {
	for i, u := range c.accessOrder {
		if u == url {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			c.accessOrder = append(c.accessOrder, url)
			return
		}
	}
}

func (c *InMemoryCache) removeFromAccessOrder(url string) {
	for i, u := range c.accessOrder {
		if u == url {
			c.accessOrder = append(c.accessOrder[:i], c.accessOrder[i+1:]...)
			return
		}
	}
}

func (c *InMemoryCache) evictLRU() {
	if len(c.accessOrder) == 0 {
		return
	}
	lru := c.accessOrder[0]
	delete(c.entries, lru)
	c.accessOrder = c.accessOrder[1:]
}
