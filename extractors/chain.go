package extractors

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/aluknot/scraper-lib/types"
)

// ErrAllExtractorsFailed is returned when every extractor in the chain fails.
type ErrAllExtractorsFailed struct {
	URL      string
	Attempts []types.Attempt
}

func (e ErrAllExtractorsFailed) Error() string {
	return "all extractors failed for " + e.URL
}

// Chain holds an ordered list of extractors and runs them in priority order
// until one succeeds with valid content.
type Chain struct {
	extractors []Extractor
}

// NewChain creates a new Chain with extractors sorted by priority.
func NewChain(extractors ...Extractor) *Chain {
	sort.Slice(extractors, func(i, j int) bool {
		return extractors[i].Priority() < extractors[j].Priority()
	})
	return &Chain{extractors: extractors}
}

// Extract runs each extractor in priority order. Returns the first result
// that passes IsValid(). If all fail, returns an empty result with
// ErrAllExtractorsFailed.
func (c *Chain) Extract(ctx context.Context, htmlContent, url string) (*types.ExtractResult, error) {
	var attempts []types.Attempt

	for _, e := range c.extractors {
		start := time.Now()
		result, err := e.Extract(ctx, htmlContent, url)
		elapsed := time.Since(start).Milliseconds()

		attempt := types.Attempt{
			Extractor:  e.Name(),
			DurationMs: elapsed,
		}

		if err != nil {
			attempt.Status = "error"
			attempt.Error = err.Error()
			attempts = append(attempts, attempt)
			slog.Debug("extractor_error",
				"url", url,
				"extractor", e.Name(),
				"status", "error",
				"error", err.Error(),
				"duration_ms", elapsed)
			continue
		}

		if result == nil {
			attempt.Status = "error"
			attempt.Error = "nil result"
			attempts = append(attempts, attempt)
			slog.Debug("extractor_error",
				"url", url,
				"extractor", e.Name(),
				"status", "error",
				"error", "nil result",
				"duration_ms", elapsed)
			continue
		}

		if !result.IsValid() {
			attempt.Status = "low_quality"
			attempt.Error = "word_count < 100"
			attempts = append(attempts, attempt)
			slog.Debug("extractor_low_quality",
				"url", url,
				"extractor", e.Name(),
				"word_count", result.WordCount,
				"content_length", len(result.Content),
				"duration_ms", elapsed)
			continue
		}

		attempt.Status = "success"
		attempts = append(attempts, attempt)
		result.Attempts = attempts
		slog.Debug("extractor_success",
			"url", url,
			"extractor", e.Name(),
			"word_count", result.WordCount,
			"content_length", len(result.Content),
			"duration_ms", elapsed)
		return result, nil
	}

	slog.Debug("all_extractors_failed",
		"url", url,
		"attempts", len(attempts),
		"attempt_summary", func() string {
			summary := ""
			for _, a := range attempts {
				summary += a.Extractor + ":" + a.Status + " "
			}
			return summary
		}())

	return &types.ExtractResult{Attempts: attempts}, ErrAllExtractorsFailed{URL: url, Attempts: attempts}
}

// DefaultChain creates a standard extraction chain with the built-in
// extractors in the recommended order:
//
//	0 — domain-specific (knows when it doesn't apply)
//	1 — readability (fastest general-purpose)
//	2 — trafilatura (better accuracy, slower)
//	3 — fallback (last resort, simple selectors)
func DefaultChain() *Chain {
	return NewChain(
		NewDomainSpecificExtractor(), // priority 0
		NewReadabilityExtractor(),    // priority 1
		NewTrafilaturaExtractor(),    // priority 2
		NewFallbackExtractor(),       // priority 3
	)
}
