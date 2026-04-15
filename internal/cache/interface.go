// Package cache provides caching interfaces and implementations for scraper-lib.
package cache

import (
	"time"
)

// Cache is the interface that all cache implementations must satisfy.
// Implementations must be safe for concurrent use.
type Cache interface {
	// Get retrieves a cached result. Returns (result, true) on hit, (nil, false) on miss.
	Get(url string) (*Result, bool)

	// Set stores a result in the cache with the given TTL.
	Set(url string, result *Result, ttl time.Duration)

	// Delete removes a specific URL from the cache.
	Delete(url string) error

	// Clear removes all entries from the cache.
	Clear() error

	// Stats returns cache statistics.
	Stats() Stats
}

// Result holds the cached data for a URL extraction.
type Result struct {
	URL       string    `json:"url"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Language  string    `json:"language"`
	Extractor string    `json:"extractor"`
	WordCount int       `json:"word_count"`
	Warnings  []string  `json:"warnings"`
	FetchedAt time.Time `json:"fetched_at"`
}

// Stats holds cache statistics.
type Stats struct {
	Hits   int `json:"hits"`
	Misses int `json:"misses"`
	Size   int `json:"size"`
}
