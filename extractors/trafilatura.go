package extractors

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aluknot/scraper-lib/sanitize"
	"github.com/aluknot/scraper-lib/types"
	"github.com/markusmobius/go-trafilatura"
	"golang.org/x/net/html"
)

// TrafilaturaExtractor extracts article content using go-trafilatura.
// Slower than readability but often more accurate on complex pages.
type TrafilaturaExtractor struct {
	sanitizer *sanitize.Sanitizer
}

// NewTrafilaturaExtractor creates a new TrafilaturaExtractor.
func NewTrafilaturaExtractor() *TrafilaturaExtractor {
	return &TrafilaturaExtractor{
		sanitizer: sanitize.NewSanitizer(),
	}
}

func (e *TrafilaturaExtractor) Name() string  { return "trafilatura" }
func (e *TrafilaturaExtractor) Priority() int { return 2 }

func (e *TrafilaturaExtractor) Extract(ctx context.Context, htmlContent, articleURL string) (*types.ExtractResult, error) {
	parsedURL, err := url.Parse(articleURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	opts := trafilatura.Options{
		OriginalURL:     parsedURL,
		EnableFallback:  true,
		TargetLanguage:  "",
		ExcludeComments: true,
	}

	result, err := trafilatura.Extract(strings.NewReader(htmlContent), opts)
	if err != nil {
		return nil, fmt.Errorf("trafilatura extract: %w", err)
	}

	if result == nil || result.ContentNode == nil {
		return nil, fmt.Errorf("trafilatura: no content extracted")
	}

	htmlContentNode, err := nodeToHTML(result.ContentNode)
	if err != nil {
		return nil, fmt.Errorf("trafilatura render: %w", err)
	}
	htmlContentNode = stripBodyWrapper(htmlContentNode)

	extracted := &types.ExtractResult{
		Content:       e.sanitizer.Clean(htmlContentNode),
		ExtractorUsed: "trafilatura",
	}

	if result.Metadata.Title != "" {
		extracted.Title = result.Metadata.Title
	}
	if result.Metadata.Author != "" {
		extracted.Author = result.Metadata.Author
	}
	if result.Metadata.Description != "" {
		// Used as summary, not stored in ExtractResult yet
	}
	if result.Metadata.Language != "" {
		extracted.Language = result.Metadata.Language
	}
	if !result.Metadata.Date.IsZero() {
		extracted.PublishedAt = &result.Metadata.Date
	}

	extracted.WordCount = countWords(extracted.Content)

	return extracted, nil
}

// nodeToHTML renders an *html.Node to its HTML string representation.
func nodeToHTML(node *html.Node) (string, error) {
	var buf strings.Builder
	if err := html.Render(&buf, node); err != nil {
		return "", fmt.Errorf("render html node: %w", err)
	}
	return buf.String(), nil
}

// stripBodyWrapper removes <body> wrapper tags that trafilatura sometimes adds.
func stripBodyWrapper(htmlContent string) string {
	htmlContent = strings.TrimPrefix(htmlContent, "<body>")
	htmlContent = strings.TrimPrefix(htmlContent, "<body ")
	// Handle <body ...> with attributes
	if idx := strings.Index(htmlContent, "<body"); idx == 0 {
		if endIdx := strings.Index(htmlContent, ">"); endIdx > 0 {
			htmlContent = htmlContent[endIdx+1:]
		}
	}
	htmlContent = strings.TrimSuffix(htmlContent, "</body>")
	return strings.TrimSpace(htmlContent)
}
