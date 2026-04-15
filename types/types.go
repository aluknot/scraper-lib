// Package types defines shared data structures used across scraper-lib.
package types

import "time"

// Attempt records the outcome of a single extractor execution.
type Attempt struct {
	Extractor  string `json:"extractor"`
	Status     string `json:"status"` // success, error, low_quality
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// StrategyAttempt records the outcome of a single fetch strategy execution.
type StrategyAttempt struct {
	Strategy   string `json:"strategy"`
	Status     string `json:"status"` // success, error, escalated, blocked
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

// ExtractResult holds the raw output from an extractor before output formatting.
type ExtractResult struct {
	Content     string
	Title       string
	Author      string
	PublishedAt *time.Time
	Language    string
	Images      []string
	Videos      []string // URLs of videos found (YouTube, Vimeo)
	Links       []string // URLs of links found in content
	Category    string   // Auto-detected content type: "song", "tutorial", "repo", etc.

	// Metadata específica para outputs especializados
	Price *PriceInfo
	Job   *JobInfo

	// Diagnóstico — siempre presente
	ExtractorUsed string
	QualityScore  float64
	WordCount     int
	Attempts      []Attempt
	Warnings      []string
}

// IsValid reports whether the extraction result meets minimum quality thresholds.
func (r *ExtractResult) IsValid() bool {
	return r.WordCount >= 100 && r.Content != ""
}

// PriceInfo holds structured product pricing data extracted from e-commerce pages.
type PriceInfo struct {
	ProductName  string
	Price        float64
	Currency     string
	Availability string // in_stock, out_of_stock, preorder
	Brand        string
	SKU          string
}

// JobInfo holds structured job listing data extracted from job posting pages.
type JobInfo struct {
	Company      string
	Location     string
	Salary       string
	JobType      string // full_time, part_time, contract
	Remote       bool
	Requirements []string
	Benefits     []string
	ApplyURL     string
}
