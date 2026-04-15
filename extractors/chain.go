package extractors

import (
	"context"
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
		attempt := types.Attempt{
			Extractor:  e.Name(),
			DurationMs: time.Since(start).Milliseconds(),
		}

		if err != nil {
			attempt.Status = "error"
			attempt.Error = err.Error()
			attempts = append(attempts, attempt)
			continue
		}

		if result == nil || !result.IsValid() {
			attempt.Status = "low_quality"
			attempts = append(attempts, attempt)
			continue
		}

		attempt.Status = "success"
		attempts = append(attempts, attempt)
		result.Attempts = attempts
		return result, nil
	}

	return &types.ExtractResult{Attempts: attempts}, ErrAllExtractorsFailed{URL: url, Attempts: attempts}
}

// DefaultChain creates a standard extraction chain with the built-in
// extractors in the recommended order:
//
//	0 — domain-specific (knows when it doesn't apply)
//	1 — readability (fastest general-purpose)
//	2 — trafilatura (better accuracy, slower)
//	3 — colly (last resort, simple selectors)
func DefaultChain() *Chain {
	return NewChain(
		NewDomainSpecificExtractor(), // priority 0
		NewReadabilityExtractor(),    // priority 1
		NewTrafilaturaExtractor(),    // priority 2
		NewCollyExtractor(),          // priority 3
	)
}
