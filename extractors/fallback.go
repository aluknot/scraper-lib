// Package extractors provides content extraction implementations for various sources.
//
// The package contains two main categories:
//
//   - Generic extractors (in this directory): Use CSS selectors and algorithms
//     to extract content from traditional websites (blogs, news articles).
//     See: readability.go, trafilatura.go, fallback.go
//
//   - Platform extractors (in platforms/): Specialized extractors for specific
//     platforms like YouTube, GitHub, X/Twitter. These can extract both metadata
//     and content specific to each platform.
//     See: platforms/youtube/, platforms/github/
package extractors

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/types"
)

// FallbackExtractor uses simple CSS selectors as a last-resort extraction method.
// It does not extract metadata — only content text.
// It parses the already-fetched HTML in memory (no additional HTTP requests).
type FallbackExtractor struct{}

// NewFallbackExtractor creates a new FallbackExtractor.
func NewFallbackExtractor() *FallbackExtractor {
	return &FallbackExtractor{}
}

func (e *FallbackExtractor) Name() string  { return "fallback" }
func (e *FallbackExtractor) Priority() int { return 3 }

func (e *FallbackExtractor) Extract(ctx context.Context, htmlContent, articleURL string) (*types.ExtractResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("fallback parse html: %w", err)
	}

	var scrapedContent strings.Builder

	// Primary selector: <article>
	doc.Find("article").Each(func(_ int, sel *goquery.Selection) {
		text := cleanText(sel.Text())
		if text != "" {
			scrapedContent.WriteString(text)
			scrapedContent.WriteString("\n\n")
		}
	})

	// Fallback: [class*='article']
	if scrapedContent.Len() < 100 {
		doc.Find("[class*='article']").Each(func(_ int, sel *goquery.Selection) {
			text := cleanText(sel.Text())
			if len(text) > 200 {
				scrapedContent.WriteString(text)
				scrapedContent.WriteString("\n\n")
			}
		})
	}

	// Fallback: <main>
	if scrapedContent.Len() < 100 {
		doc.Find("main").Each(func(_ int, sel *goquery.Selection) {
			text := cleanText(sel.Text())
			if len(text) > 200 {
				scrapedContent.WriteString(text)
			}
		})
	}

	// Fallback: div[class*='content']
	if scrapedContent.Len() < 100 {
		doc.Find("div[class*='content']").Each(func(_ int, sel *goquery.Selection) {
			text := cleanText(sel.Text())
			if len(text) > 300 {
				scrapedContent.WriteString(text)
			}
		})
	}

	content := strings.TrimSpace(scrapedContent.String())
	if content == "" {
		return nil, fmt.Errorf("fallback: no content extracted")
	}

	return &types.ExtractResult{
		Content:       content,
		ExtractorUsed: "fallback",
		WordCount:     countWords(content),
	}, nil
}

// cleanText sanitizes raw text extracted from HTML elements.
func cleanText(text string) string {
	lines := strings.Split(text, "\n")
	var cleanLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 10 {
			cleanLines = append(cleanLines, line)
		}
	}

	result := strings.Join(cleanLines, "\n")
	result = strings.ReplaceAll(result, "\t", " ")
	for strings.Contains(result, "  ") {
		result = strings.ReplaceAll(result, "  ", " ")
	}

	return strings.TrimSpace(result)
}
