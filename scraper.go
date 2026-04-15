// Package scraperlib provides a high-level API for fetching and extracting
// article content from web pages.
package scraperlib

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aluknot/scraper-lib/embeds"
	"github.com/aluknot/scraper-lib/internal/cache"
	"github.com/aluknot/scraper-lib/internal/detection"
	"github.com/aluknot/scraper-lib/internal/fetch"
	"github.com/aluknot/scraper-lib/sanitize"
	"github.com/aluknot/scraper-lib/types"

	"github.com/aluknot/scraper-lib/extractors"
	"github.com/aluknot/scraper-lib/output"
)

// DefaultCacheTTL is the default time-to-live for cached results.
const DefaultCacheTTL = 24 * time.Hour

// defaultCache is the shared in-memory cache used when no custom cache is provided.
var defaultCache = cache.NewInMemoryCache()

// Options configures the extraction behavior.
type Options struct {
	Timeout       time.Duration
	UserAgent     string
	ExtractImages bool
	Outputs       []string // Preferred: allows multiple outputs ["article", "metadata", "price"]
	Output        string   // DEPRECATED: use Outputs. Kept for backwards compat.

	// HTTP strategy — when UseAdvanced is true, uses UA rotation, cookie jar,
	// and referrer spoofing. Defaults to simple http.Client.
	UseAdvanced bool

	// AdvancedHTTP configures the advanced fetcher when UseAdvanced is true.
	// When nil, defaults are used (UA rotation + referrer spoofing enabled).
	AdvancedHTTP *fetch.AdvancedOptions

	// Extractor override — mutually exclusive, first one set wins.

	// Extractors replaces the entire extractor chain.
	// When set, DefaultChain() is not used. Order matters.
	Extractors []extractors.Extractor

	// Extractor forces a single extractor by name.
	// Valid names: "domain_specific", "readability", "trafilatura", "colly".
	// When set, Extractors and DefaultChain are ignored.
	Extractor string

	// NoFallback prevents falling back to the next extractor in the chain.
	// If the first (or forced) extractor fails, an error is returned
	// immediately instead of trying the next one.
	NoFallback bool

	// Cache overrides the default in-memory cache. Set to nil to disable caching.
	Cache cache.Cache

	// CacheTTL sets how long results are cached. Defaults to 24 hours.
	CacheTTL time.Duration

	// DisableCache disables caching entirely. When true, Cache is ignored.
	DisableCache bool

	// Pipeline stages — default all enabled (current behavior).

	// NoSanitize skips HTML sanitization (bluemonday UGCPolicy).
	// Useful when you need raw extracted content.
	NoSanitize bool

	// NoEmbeds skips embed extraction and restoration.
	// Embeds (YouTube, Twitter, etc.) will be treated like any other HTML.
	NoEmbeds bool

	// NoPaywallDetection skips paywall signal detection.
	NoPaywallDetection bool

	// Markdown output options

	// Category overrides auto-detected category for template selection.
	// E.g., "tutorial", "song", "repo".
	Category string

	// TemplateDir is a directory of custom YAML templates.
	// If empty, uses built-in defaults.
	TemplateDir string
}

// Result holds the output of an extraction operation.
// Only one specialized output will be populated based on Options.Outputs.
// Diagnostics are always present.
type Result struct {
	// Output — one or more will be populated based on Options.Outputs
	Article  *output.ArticleResult
	Metadata *output.MetadataResult
	Raw      *output.RawResult
	Markdown *output.MarkdownResult

	// Diagnostics — always present, independent of output type
	ExtractorUsed     string          `json:"extractor_used"`
	StrategyUsed      string          `json:"strategy_used"`
	ExtractorAttempts []types.Attempt `json:"extractor_attempts"`
	QualityScore      float64         `json:"quality_score"`
	WordCount         int             `json:"word_count"`
	Warnings          []string        `json:"warnings"`
	DurationMs        int64           `json:"duration_ms"`
}

// Extract downloads and extracts article content from a URL.
func Extract(ctx context.Context, url string, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}
	opts.Normalize()

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	slog.Debug("extract_start", "url", url)

	// Resolve cache (respects DisableCache)
	var cacheInstance cache.Cache
	var cacheTTL time.Duration

	if !opts.DisableCache {
		cacheInstance = resolveCache(opts.Cache)
		cacheTTL = opts.CacheTTL
		if cacheTTL == 0 {
			cacheTTL = DefaultCacheTTL
		}
		// Check cache hit
		if result, ok := cacheInstance.Get(url); ok {
			slog.Info("cache_hit", "url", url, "duration_ms", time.Since(start).Milliseconds())
			return resultFromCache(result), nil
		}
		slog.Debug("cache_miss", "url", url)
	}

	// Resolve HTTP client — simple or advanced
	client := resolveHTTPClient(opts, timeout)
	chain := buildChain(opts)

	// Step 1: Fetch raw HTML
	var rawHTML string
	var err error
	if opts.UseAdvanced {
		slog.Debug("fetch_start", "url", url, "strategy", "advanced")
		advOpts := resolveAdvancedOptions(opts)
		rawHTML, err = fetch.GetHTMLAdvanced(ctx, client, url, advOpts)
	} else {
		slog.Debug("fetch_start", "url", url, "strategy", "simple")
		rawHTML, err = fetch.GetHTML(ctx, client, url)
	}
	if err != nil {
		slog.Error("fetch_failed", "url", url, "error", err)
		return nil, fmt.Errorf("fetch HTML: %w", err)
	}
	slog.Debug("fetch_success", "url", url, "html_size", len(rawHTML), "duration_ms", time.Since(start).Milliseconds())

	// Step 2: Detect paywall (optional)
	warnings := make([]string, 0)
	if !opts.NoPaywallDetection {
		if isPaywall, signals := detection.DetectPaywall(rawHTML); isPaywall {
			for _, s := range signals {
				warnings = append(warnings, fmt.Sprintf("paywall_detected:%s", s))
			}
			slog.Warn("paywall_detected", "url", url, "signals", signals)
		}
	}

	// Step 3: PRE — Extract embeds and replace with placeholders (optional)
	processedHTML := rawHTML
	var embedMap map[string]string
	if !opts.NoEmbeds {
		embedExtractor := embeds.NewEmbedExtractor()
		processedHTML, embedMap = embedExtractor.ExtractAndReplace(rawHTML)
		if len(embedMap) > 0 {
			slog.Debug("embeds_extracted", "url", url, "count", len(embedMap))
		}
	}

	// Step 4: Run extractor chain (readability / trafilatura / colly)
	extracted, err := chain.Extract(ctx, processedHTML, url)
	if err != nil {
		slog.Error("extraction_failed", "url", url, "error", err)
		return nil, fmt.Errorf("extract: %w", err)
	}

	// Step 5: Sanitize (optional)
	if !opts.NoSanitize {
		sanitizer := sanitize.NewSanitizer()
		extracted.Content = sanitizer.Clean(extracted.Content)
	}

	// Step 6: POST — Restore embeds (optional)
	if !opts.NoEmbeds && len(embedMap) > 0 {
		embedExtractor := embeds.NewEmbedExtractor()
		extracted.Content = embedExtractor.Restore(extracted.Content, embedMap)
	}

	// Propagate warnings from extraction
	warnings = append(warnings, extracted.Warnings...)

	result := buildResult(extracted, opts.Outputs, url, opts.Category, opts.TemplateDir, warnings, time.Since(start).Milliseconds())

	// Store in cache if enabled
	if !opts.DisableCache {
		cacheInstance.Set(url, toCacheResult(result), cacheTTL)
	}

	slog.Info("extraction_complete",
		"url", url,
		"extractor", result.ExtractorUsed,
		"word_count", result.WordCount,
		"duration_ms", time.Since(start).Milliseconds(),
		"warnings", len(warnings))

	return result, nil
}

// ExtractHTML extracts article content from already-downloaded HTML.
// Useful for testing and when the caller already has the HTML.
func ExtractHTML(ctx context.Context, htmlContent string, baseURL string, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}
	opts.Normalize()

	// Resolve cache (respects DisableCache)
	var cacheInstance cache.Cache
	var cacheTTL time.Duration

	if !opts.DisableCache {
		cacheInstance = resolveCache(opts.Cache)
		cacheTTL = opts.CacheTTL
		if cacheTTL == 0 {
			cacheTTL = DefaultCacheTTL
		}
		// Check cache hit
		if result, ok := cacheInstance.Get(baseURL); ok {
			return resultFromCache(result), nil
		}
	}

	chain := buildChain(opts)

	start := time.Now()

	// Detect paywall (optional)
	warnings := make([]string, 0)
	if !opts.NoPaywallDetection {
		if isPaywall, signals := detection.DetectPaywall(htmlContent); isPaywall {
			for _, s := range signals {
				warnings = append(warnings, fmt.Sprintf("paywall_detected:%s", s))
			}
		}
	}

	// Extract embeds (optional)
	processedHTML := htmlContent
	var embedMap map[string]string
	if !opts.NoEmbeds {
		embedExtractor := embeds.NewEmbedExtractor()
		processedHTML, embedMap = embedExtractor.ExtractAndReplace(htmlContent)
	}

	// Run extractor chain
	extracted, err := chain.Extract(ctx, processedHTML, baseURL)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	// Sanitize before restoring (optional)
	if !opts.NoSanitize {
		sanitizer := sanitize.NewSanitizer()
		extracted.Content = sanitizer.Clean(extracted.Content)
	}

	// Restore embeds (optional)
	if !opts.NoEmbeds && len(embedMap) > 0 {
		embedExtractor := embeds.NewEmbedExtractor()
		extracted.Content = embedExtractor.Restore(extracted.Content, embedMap)
	}

	warnings = append(warnings, extracted.Warnings...)

	result := buildResult(extracted, opts.Outputs, baseURL, opts.Category, opts.TemplateDir, warnings, time.Since(start).Milliseconds())

	// Store in cache if enabled
	if !opts.DisableCache {
		cacheInstance.Set(baseURL, toCacheResult(result), cacheTTL)
	}

	return result, nil
}

// Normalize ensures Outputs has at least one value.
// Backwards compat: if Output is set but Outputs is not, uses Output.
func (o *Options) Normalize() {
	if len(o.Outputs) == 0 && o.Output != "" {
		o.Outputs = []string{o.Output}
	}
	if len(o.Outputs) == 0 {
		o.Outputs = []string{"article"} // default
	}
}

// buildChain creates an extractor chain based on Options settings.
// Priority: Extractor (by name) > Extractors (custom list) > DefaultChain().
// If NoFallback is true, only the first extractor is used.
func buildChain(opts *Options) *extractors.Chain {
	// Force single extractor by name — inherently no fallback
	if opts.Extractor != "" {
		e := extractorByName(opts.Extractor)
		if e != nil {
			return extractors.NewChain(e)
		}
		// Unknown name falls back to default chain
	}

	// Custom chain
	if len(opts.Extractors) > 0 {
		if opts.NoFallback {
			return extractors.NewChain(opts.Extractors[0])
		}
		return extractors.NewChain(opts.Extractors...)
	}

	// Default chain — respect NoFallback
	if opts.NoFallback {
		return extractors.NewChain(
			extractors.NewDomainSpecificExtractor(),
		)
	}

	return extractors.DefaultChain()
}

// extractorByName returns an Extractor instance by its name string.
func extractorByName(name string) extractors.Extractor {
	switch name {
	case "domain_specific":
		return extractors.NewDomainSpecificExtractor()
	case "readability":
		return extractors.NewReadabilityExtractor()
	case "trafilatura":
		return extractors.NewTrafilaturaExtractor()
	case "colly":
		return extractors.NewCollyExtractor()
	default:
		return nil
	}
}

func buildResult(extracted *types.ExtractResult, outputTypes []string, url, category, templateDir string, warnings []string, durationMs int64) *Result {
	result := &Result{
		ExtractorUsed:     extracted.ExtractorUsed,
		ExtractorAttempts: extracted.Attempts,
		QualityScore:      extracted.QualityScore,
		WordCount:         extracted.WordCount,
		Warnings:          warnings,
		DurationMs:        durationMs,
	}

	for _, outputType := range outputTypes {
		switch outputType {
		case "article":
			result.Article = output.BuildArticleResult(extracted, url)
		case "metadata":
			result.Metadata = output.BuildMetadataResult(extracted, url)
		case "raw":
			result.Raw = output.BuildRawResult(extracted)
		case "markdown":
			md, err := output.BuildMarkdownResult(extracted, url, category, templateDir)
			if err == nil {
				result.Markdown = md
			}
		}
	}

	// Default to article if no output matched
	if result.Article == nil && result.Metadata == nil && result.Raw == nil {
		result.Article = output.BuildArticleResult(extracted, url)
	}

	return result
}

// resolveCache returns the provided cache, or the shared default InMemoryCache if nil.
func resolveCache(c cache.Cache) cache.Cache {
	if c == nil {
		return defaultCache
	}
	return c
}

// toCacheResult converts a scraperlib.Result to a cache.Result for storage.
func toCacheResult(r *Result) *cache.Result {
	cr := &cache.Result{
		Title:     "",
		Author:    "",
		Language:  "",
		Extractor: r.ExtractorUsed,
		WordCount: r.WordCount,
		Warnings:  r.Warnings,
		FetchedAt: time.Now(),
	}
	if r.Article != nil {
		cr.URL = r.Article.URL
		cr.Title = r.Article.Title
		cr.Author = r.Article.Author
		cr.Language = r.Article.Language
		cr.Content = r.Article.Content
	} else if r.Raw != nil {
		cr.Content = r.Raw.Content
	} else if r.Metadata != nil {
		cr.Title = r.Metadata.Title
		cr.Author = r.Metadata.Author
		cr.Language = r.Metadata.Language
		cr.WordCount = r.Metadata.WordCount
	}
	return cr
}

// resultFromCache rebuilds a scraperlib.Result from cached data.
func resultFromCache(cr *cache.Result) *Result {
	result := &Result{
		ExtractorUsed: cr.Extractor,
		WordCount:     cr.WordCount,
		Warnings:      cr.Warnings,
	}

	result.Article = &output.ArticleResult{
		Title:   cr.Title,
		Content: cr.Content,
		Author:  cr.Author,
		URL:     cr.URL,
	}
	if cr.Language != "" {
		result.Article.Language = cr.Language
	}

	return result
}

// resolveHTTPClient returns an http.Client configured per the options.
// When UseAdvanced is true, creates a client with cookie jar support.
func resolveHTTPClient(opts *Options, timeout time.Duration) *http.Client {
	if opts.UseAdvanced {
		return fetch.NewHTTPClientAdvanced(timeout, true)
	}
	return &http.Client{Timeout: timeout}
}

// resolveAdvancedOptions returns the advanced HTTP options, filling in
// sensible defaults when not explicitly set.
func resolveAdvancedOptions(opts *Options) *fetch.AdvancedOptions {
	if opts.AdvancedHTTP != nil {
		return opts.AdvancedHTTP
	}
	// Defaults: enable all advanced features
	return &fetch.AdvancedOptions{
		RotateUserAgent:     true,
		SpoofReferrer:       true,
		SpoofAcceptLanguage: true,
		EnableCookies:       true,
	}
}
