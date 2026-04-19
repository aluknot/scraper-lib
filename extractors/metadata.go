package extractors

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/types"
)

// MetadataExtractor extracts just the metadata (title, description) from HTML.
// This is an ultra-lightweight extractor that does NOT use readability/trafilatura.
// It only parses meta tags - perfect for quick metadata extraction or as a fallback
// when full content extraction is not needed.
type MetadataExtractor struct{}

// NewMetadataExtractor creates a new MetadataExtractor.
func NewMetadataExtractor() *MetadataExtractor {
	return &MetadataExtractor{}
}

func (e *MetadataExtractor) Name() string  { return "metadata" }
func (e *MetadataExtractor) Priority() int { return 0 }

// Extract extracts only metadata from HTML.
// It does NOT extract article content - just meta tags.
func (e *MetadataExtractor) Extract(ctx context.Context, htmlContent, pageURL string) (*types.ExtractResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	result := &types.ExtractResult{
		ExtractorUsed: "metadata",
		WordCount:     0,
		Content:       "", // No article content
	}

	// Title - priority: og:title > title > twitter:title
	title, exists := doc.Find("meta[property='og:title']").Attr("content")
	if !exists || title == "" {
		title = doc.Find("meta[name='twitter:title']").AttrOr("content", "")
	}
	if title == "" {
		title = doc.Find("title").First().Text()
	}
	result.Title = strings.TrimSpace(title)

	// Description - priority: og:description > description
	desc, exists := doc.Find("meta[property='og:description']").Attr("content")
	if !exists || desc == "" {
		desc = doc.Find("meta[name='description']").AttrOr("content", "")
	}
	result.Title = strings.TrimSpace(result.Title)

	// Author
	author, _ := doc.Find("meta[name='author']").Attr("content")
	if author == "" {
		author, _ = doc.Find("meta[property='article:author']").Attr("content")
	}
	result.Author = strings.TrimSpace(author)

	// Language
	lang, _ := doc.Find("html").Attr("lang")
	if lang == "" {
		lang, _ = doc.Find("meta[http-equiv='content-language']").Attr("content")
	}
	result.Language = strings.TrimSpace(lang)

	// Published date
	publishedStr, _ := doc.Find("meta[property='article:published_time']").Attr("content")
	if publishedStr != "" {
		if t, err := parseRFC3339(publishedStr); err == nil {
			result.PublishedAt = t
		}
	}

	// Site name
	siteName, _ := doc.Find("meta[property='og:site_name']").Attr("content")
	if siteName == "" {
		parsedURL, _ := url.Parse(pageURL)
		if parsedURL != nil {
			siteName = parsedURL.Host
		}
	}

	// Images
	doc.Find("meta[property='og:image']").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("content")
		if src != "" {
			result.Images = append(result.Images, src)
		}
	})

	// Videos
	doc.Find("meta[property='og:video:url']").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("content")
		if src != "" {
			result.Videos = append(result.Videos, src)
		}
	})
	// Also check for embedded videos
	doc.Find("meta[property='og:video']").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("content")
		if src != "" {
			result.Videos = append(result.Videos, src)
		}
	})

	// Links found in content - for metadata mode we just return empty
	// The content field is intentionally empty for this extractor

	// Category from og:type
	category, _ := doc.Find("meta[property='og:type']").Attr("content")
	result.Category = strings.TrimSpace(category)

	return result, nil
}

// Helper function to parse RFC3339 date string.
func parseRFC3339(dateStr string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
