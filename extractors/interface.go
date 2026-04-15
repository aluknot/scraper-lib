// Package extractors defines the Extractor interface and shared types.
package extractors

import (
	"context"

	"github.com/aluknot/scraper-lib/types"
)

// Extractor is the interface that all content extractors must implement.
type Extractor interface {
	// Name returns the human-readable name of the extractor.
	Name() string

	// Priority returns the execution order. Lower values run first.
	Priority() int

	// Extract attempts to extract article content from the given HTML.
	// Returns an error if the extractor cannot process the HTML at all.
	// Returns a result with IsValid() == false if extraction produced
	// insufficient content (not an error, just low quality).
	Extract(ctx context.Context, html string, url string) (*types.ExtractResult, error)
}
