package extractors

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"github.com/aluknot/scraper-lib/sanitize"
	"github.com/aluknot/scraper-lib/types"
)

// ReadabilityExtractor extracts article content using go-readability v2.
// This is the fastest general-purpose extractor and is preferred for most sites.
type ReadabilityExtractor struct {
	sanitizer *sanitize.Sanitizer
}

// NewReadabilityExtractor creates a new ReadabilityExtractor.
func NewReadabilityExtractor() *ReadabilityExtractor {
	return &ReadabilityExtractor{
		sanitizer: sanitize.NewSanitizer(),
	}
}

func (e *ReadabilityExtractor) Name() string  { return "readability" }
func (e *ReadabilityExtractor) Priority() int { return 1 }

func (e *ReadabilityExtractor) Extract(ctx context.Context, htmlContent, articleURL string) (*types.ExtractResult, error) {
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		return nil, fmt.Errorf("readability: %w", err)
	}

	if article.Node == nil {
		return nil, fmt.Errorf("readability: no article content found")
	}

	var buf strings.Builder
	if err := article.RenderHTML(&buf); err != nil {
		return nil, fmt.Errorf("readability render: %w", err)
	}
	content := e.sanitizer.Clean(strings.TrimSpace(buf.String()))

	result := &types.ExtractResult{
		Content:       content,
		ExtractorUsed: "readability",
		WordCount:     countWords(content),
	}

	result.Title = article.Title()
	if result.Title == "" {
		result.Title = article.SiteName()
	}
	result.Author = article.Byline()
	// Note: article.Excerpt() returns summary, not full content
	result.Language = article.Language()

	if t, err := article.PublishedTime(); err == nil && !t.IsZero() {
		result.PublishedAt = &t
	}

	return result, nil
}

// countWords returns the approximate number of visible words in the HTML content.
// Strips HTML tags and splits on whitespace.
func countWords(htmlContent string) int {
	// Simple approach: strip tags, count whitespace-separated tokens
	var inTag bool
	var textBuf strings.Builder
	for _, r := range htmlContent {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				textBuf.WriteRune(r)
			}
		}
	}

	text := strings.TrimSpace(textBuf.String())
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	return len(words)
}
