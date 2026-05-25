// Package cache provides caching interfaces and implementations for scraper-lib.
package cache

import (
	"encoding/json"
	"time"

	"github.com/aluknot/scraper-lib/extractors/platforms/youtube"
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
	URL          string    `json:"url"`
	Content      string    `json:"content"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	Language     string    `json:"language"`
	Extractor    string    `json:"extractor"`
	WordCount    int       `json:"word_count"`
	Warnings     []string  `json:"warnings"`
	FetchedAt    time.Time `json:"fetched_at"`
	Description  string    `json:"description"`
	SiteName     string    `json:"site_name"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Category     string    `json:"category"`
	PlatformData struct {
		YouTube *youtube.VideoMetadata `json:"youtube,omitempty"`
	} `json:"platform_data,omitempty"`
}

// Stats holds cache statistics.
type Stats struct {
	Hits   int `json:"hits"`
	Misses int `json:"misses"`
	Size   int `json:"size"`
}

// MarshalJSON implements custom JSON marshaling to handle VideoMetadata.
func (r *Result) MarshalJSON() ([]byte, error) {
	type Alias Result
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling.
func (r *Result) UnmarshalJSON(data []byte) error {
	type Alias Result
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	return json.Unmarshal(data, aux)
}
