package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileCache stores cache entries as JSON files in a directory.
type FileCache struct {
	dir    string
	mu     sync.Mutex
	hits   int
	misses int
}

// fileEntry wraps a cache Result with metadata for file storage.
type fileEntry struct {
	Result    *Result `json:"result"`
	TTLMillis int64   `json:"ttl_millis"`
	CreatedAt string  `json:"created_at"`
}

// NewFileCache creates a file-based cache in the given directory.
// Creates the directory if it doesn't exist.
func NewFileCache(dir string) (*FileCache, error) {
	if dir == "" {
		dir = defaultCacheDir()
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}
	return &FileCache{dir: dir}, nil
}

// Get retrieves a cached result from a JSON file.
func (c *FileCache) Get(url string) (*Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.entryPath(url)
	data, err := os.ReadFile(path)
	if err != nil {
		c.misses++
		return nil, false
	}

	var entry fileEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		c.misses++
		return nil, false
	}

	if entry.Result == nil {
		c.misses++
		return nil, false
	}

	createdAt, err := time.Parse(time.RFC3339, entry.CreatedAt)
	if err != nil {
		c.misses++
		return nil, false
	}

	elapsed := time.Since(createdAt)
	ttl := time.Duration(entry.TTLMillis) * time.Millisecond
	if elapsed > ttl {
		os.Remove(path)
		c.misses++
		return nil, false
	}

	c.hits++
	return entry.Result, true
}

// Set stores a result as a JSON file in the cache directory.
func (c *FileCache) Set(url string, result *Result, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := fileEntry{
		Result:    result,
		TTLMillis: ttl.Milliseconds(),
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return // Silently fail — cache is non-critical
	}

	path := c.entryPath(url)
	// Write directly to the file — atomic enough for cache purposes
	if err := os.WriteFile(path, data, 0644); err != nil {
		return
	}
}

// Delete removes a specific URL's cache file.
func (c *FileCache) Delete(url string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.entryPath(url)
	return os.Remove(path)
}

// Clear removes all cache files from the directory.
func (c *FileCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return fmt.Errorf("read cache dir: %w", err)
	}

	for _, e := range entries {
		if e.Name() == "." || e.Name() == ".." {
			continue
		}
		os.Remove(filepath.Join(c.dir, e.Name()))
	}

	c.hits = 0
	c.misses = 0
	return nil
}

// Stats returns cache statistics.
func (c *FileCache) Stats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()

	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return Stats{Hits: c.hits, Misses: c.misses}
	}

	size := 0
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			size++
		}
	}

	return Stats{
		Hits:   c.hits,
		Misses: c.misses,
		Size:   size,
	}
}

// entryPath returns the file path for a URL's cache entry.
func (c *FileCache) entryPath(url string) string {
	hash := sha256.Sum256([]byte(url))
	return filepath.Join(c.dir, fmt.Sprintf("%x.json", hash[:]))
}

// defaultCacheDir returns the default cache directory.
func defaultCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, "scraper-lib")
}
