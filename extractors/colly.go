package extractors

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/types"
)

// CollyExtractor uses simple CSS selectors as a last-resort extraction method.
// It does not extract metadata — only content text.
// It parses the already-fetched HTML in memory (no additional HTTP requests).
type CollyExtractor struct{}

// NewCollyExtractor creates a new CollyExtractor.
func NewCollyExtractor() *CollyExtractor {
	return &CollyExtractor{}
}

func (e *CollyExtractor) Name() string  { return "colly" }
func (e *CollyExtractor) Priority() int { return 3 }

func (e *CollyExtractor) Extract(ctx context.Context, htmlContent, articleURL string) (*types.ExtractResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("colly parse html: %w", err)
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
		return nil, fmt.Errorf("colly: no content extracted")
	}

	return &types.ExtractResult{
		Content:       content,
		ExtractorUsed: "colly",
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
