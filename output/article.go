// Package output formats extracted content into various output types.
package output

import (
	"strings"
	"time"

	"github.com/aluknot/scraper-lib/internal/markdown"
	"github.com/aluknot/scraper-lib/internal/urlutil"
	"github.com/aluknot/scraper-lib/types"
)

// ArticleResult holds the formatted article output.
type ArticleResult struct {
	Title       string
	Content     string
	Author      string
	PublishedAt string
	Language    string
	URL         string
}

// MetadataResult holds extracted metadata about the page/article.
type MetadataResult struct {
	Title        string
	Author       string
	Language     string
	WordCount    int
	Images       int
	Videos       int
	Links        int
	Extractor    string
	QualityScore float64
	Warnings     []string
}

// RawResult holds the unprocessed extraction result.
type RawResult struct {
	Content   string
	Extractor string
	Attempts  []types.Attempt
	WordCount int
}

// BuildArticleResult creates a formatted article result.
func BuildArticleResult(extracted *types.ExtractResult, url string) *ArticleResult {
	result := &ArticleResult{
		Title:   extracted.Title,
		Content: extracted.Content,
		Author:  extracted.Author,
		URL:     url,
	}
	if extracted.PublishedAt != nil {
		result.PublishedAt = extracted.PublishedAt.Format("2006-01-02")
	}
	if extracted.Language != "" {
		result.Language = extracted.Language
	}
	return result
}

// BuildMetadataResult creates a metadata summary result.
func BuildMetadataResult(extracted *types.ExtractResult, url string) *MetadataResult {
	return &MetadataResult{
		Title:        extracted.Title,
		Author:       extracted.Author,
		Language:     extracted.Language,
		WordCount:    extracted.WordCount,
		Images:       len(extracted.Images),
		Videos:       len(extracted.Videos),
		Links:        len(extracted.Links),
		Extractor:    extracted.ExtractorUsed,
		QualityScore: extracted.QualityScore,
		Warnings:     extracted.Warnings,
	}
}

// BuildRawResult creates a raw, unformatted result for debugging.
func BuildRawResult(extracted *types.ExtractResult) *RawResult {
	return &RawResult{
		Content:   extracted.Content,
		Extractor: extracted.ExtractorUsed,
		Attempts:  extracted.Attempts,
		WordCount: extracted.WordCount,
	}
}

// MarkdownResult holds the rendered Markdown output.
type MarkdownResult struct {
	Content  string
	Filename string
	Tags     []string
}

// BuildMarkdownResult creates a Markdown-formatted result with YAML frontmatter.
func BuildMarkdownResult(extracted *types.ExtractResult, url string, category string, templateDir string) (*MarkdownResult, error) {
	tplMgr, err := markdown.NewManager(templateDir)
	if err != nil {
		return nil, err
	}

	// Determine extractor name for template resolution
	extractorName := extracted.ExtractorUsed
	if idx := strings.Index(extractorName, ":"); idx >= 0 {
		extractorName = extractorName[:idx]
	}

	// Auto-detect category if not set
	if category == "" {
		category = extracted.Category
	}

	tags := markdown.GenerateTags(extracted.Title, extracted.Language, 5)
	publishedAt := ""
	if extracted.PublishedAt != nil {
		publishedAt = extracted.PublishedAt.Format(time.RFC3339)
	}

	// Convert HTML content to Markdown
	content := markdown.HTMLToMarkdown(extracted.Content)

	tpl := tplMgr.Resolve(extractorName, category)

	data := markdown.TemplateData{
		Title:       extracted.Title,
		URL:         url,
		Source:      urlutil.Domain(url),
		Author:      extracted.Author,
		PublishedAt: publishedAt,
		ExtractedAt: markdown.NowRFC3339(),
		Extractor:   extractorName,
		Language:    extracted.Language,
		WordCount:   extracted.WordCount,
		Category:    category,
		Content:     content,
		Tags:        tags,
		Warnings:    extracted.Warnings,
	}

	mdContent := tplMgr.Render(tpl, data)
	filename := markdown.GenerateFilename(extracted.Title, urlutil.Domain(url))

	return &MarkdownResult{
		Content:  mdContent,
		Filename: filename,
		Tags:     tags,
	}, nil
}
