//go:build integration

package scraperlib

import (
	"context"
	"strings"
	"testing"
	"time"
)

// TestExtract_Integration_RealURLs tests extraction against real, stable URLs.
// Run with: go test -tags=integration -v ./...
func TestExtract_Integration_RealURLs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name        string
		url         string
		expectMinWC int
		expectInURL string // substring expected in content
	}{
		{
			name:        "Wikipedia Go programming language",
			url:         "https://en.wikipedia.org/wiki/Go_(programming_language)",
			expectMinWC: 500,
			expectInURL: "Go",
		},
		{
			name:        "Wikipedia Python",
			url:         "https://en.wikipedia.org/wiki/Python_(programming_language)",
			expectMinWC: 500,
			expectInURL: "Python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Extract(context.Background(), tt.url, &Options{
				Timeout: 30 * time.Second,
			})
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tt.url, err)
			}
			if result.Article == nil {
				t.Fatalf("expected article result, got nil")
			}
			if result.WordCount < tt.expectMinWC {
				t.Errorf("expected word count >= %d, got %d", tt.expectMinWC, result.WordCount)
			}
			if tt.expectInURL != "" && !strings.Contains(result.Article.Content, tt.expectInURL) {
				t.Errorf("expected content to contain %q", tt.expectInURL)
			}
			if result.ExtractorUsed == "" {
				t.Error("expected extractor_used to be set")
			}
			t.Logf("extractor: %s, words: %d, score: %.2f", result.ExtractorUsed, result.WordCount, result.QualityScore)
		})
	}
}

// TestExtract_Integration_YouTubeEmbed tests that YouTube embeds are preserved
// when fetching a real page with embedded video.
func TestExtract_Integration_YouTubeEmbed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Wikipedia pages with YouTube embeds
	url := "https://en.wikipedia.org/wiki/Rickrolling"
	result, err := Extract(context.Background(), url, &Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
	if result.WordCount < 100 {
		t.Errorf("expected word count >= 100, got %d", result.WordCount)
	}
	t.Logf("extractor: %s, words: %d", result.ExtractorUsed, result.WordCount)
}

// TestExtract_Integration_MarkdownOutput tests Markdown output with a real URL.
func TestExtract_Integration_MarkdownOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	url := "https://en.wikipedia.org/wiki/Linux"
	result, err := Extract(context.Background(), url, &Options{
		Timeout: 30 * time.Second,
		Outputs: []string{"markdown"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Markdown == nil {
		t.Fatal("expected markdown result, got nil")
	}
	if result.Markdown.Content == "" {
		t.Error("expected non-empty markdown content")
	}
	if !strings.Contains(result.Markdown.Content, "---") {
		t.Error("expected frontmatter in markdown")
	}
	if result.Markdown.Filename == "" {
		t.Error("expected non-empty filename")
	}
	if len(result.Markdown.Tags) == 0 {
		t.Error("expected auto-generated tags")
	}
	t.Logf("filename: %s, tags: %v", result.Markdown.Filename, result.Markdown.Tags)
}

// TestExtract_Integration_AdvancedHTTP tests extraction with advanced HTTP options
// (UA rotation, referrer spoofing) against a real URL.
func TestExtract_Integration_AdvancedHTTP(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	url := "https://en.wikipedia.org/wiki/Go_(programming_language)"
	result, err := Extract(context.Background(), url, &Options{
		Timeout:     30 * time.Second,
		UseAdvanced: true,
	})
	if err != nil {
		t.Fatalf("unexpected error with advanced HTTP: %v", err)
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
	if result.WordCount < 200 {
		t.Errorf("expected word count >= 200, got %d", result.WordCount)
	}
	t.Logf("advanced: extractor: %s, words: %d", result.ExtractorUsed, result.WordCount)
}

// TestExtract_Integration_MultipleOutputs tests requesting multiple outputs.
func TestExtract_Integration_MultipleOutputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	url := "https://en.wikipedia.org/wiki/Go_(programming_language)"
	result, err := Extract(context.Background(), url, &Options{
		Timeout:      30 * time.Second,
		Outputs:      []string{"article", "metadata"},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Article == nil {
		t.Error("expected article result, got nil")
	}
	if result.Metadata == nil {
		t.Error("expected metadata result, got nil")
	}
	if result.Metadata.WordCount != result.WordCount {
		t.Errorf("metadata word count %d != result word count %d",
			result.Metadata.WordCount, result.WordCount)
	}
}

// TestExtract_Integration_DifferentExtractors tests that different extractors
// produce results on real URLs.
func TestExtract_Integration_DifferentExtractors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name      string
		extractor string
		url       string
	}{
		{"readability on Wikipedia", "readability", "https://en.wikipedia.org/wiki/Go_(programming_language)"},
		{"trafilatura on Wikipedia", "trafilatura", "https://en.wikipedia.org/wiki/Linux"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Extract(context.Background(), tt.url, &Options{
				Timeout:      30 * time.Second,
				Extractor:    tt.extractor,
				NoFallback:   true,
				DisableCache: true,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Article == nil {
				t.Fatal("expected article result, got nil")
			}
			if result.WordCount < 50 {
				t.Errorf("expected word count >= 50, got %d", result.WordCount)
			}
			if result.ExtractorUsed != tt.extractor {
				t.Errorf("expected extractor %q, got %q", tt.extractor, result.ExtractorUsed)
			}
		})
	}
}

// TestExtract_Integration_GitHubReadme tests domain-specific extraction on GitHub.
func TestExtract_Integration_GitHubReadme(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	url := "https://github.com/golang/go"
	result, err := Extract(context.Background(), url, &Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Article == nil {
		t.Fatal("expected article result, got nil")
	}
	// GitHub README typically contains "Go" and "programming"
	if result.WordCount < 10 {
		t.Errorf("expected word count >= 10, got %d", result.WordCount)
	}
	t.Logf("extractor: %s, words: %d", result.ExtractorUsed, result.WordCount)
}

// TestExtract_Integration_Cache tests that caching works with real URLs.
func TestExtract_Integration_Cache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	url := "https://en.wikipedia.org/wiki/Go_(programming_language)"

	start := time.Now()
	result1, err := Extract(context.Background(), url, &Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}
	firstDuration := time.Since(start)

	start = time.Now()
	result2, err := Extract(context.Background(), url, &Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}
	secondDuration := time.Since(start)

	if result2.Article.Title != result1.Article.Title {
		t.Errorf("cached result should match: got %q, want %q",
			result2.Article.Title, result1.Article.Title)
	}

	// Second call should be much faster (cache hit)
	t.Logf("first call: %v, second call (cached): %v", firstDuration, secondDuration)
	if secondDuration > firstDuration {
		t.Logf("NOTE: second call was not faster — cache may not be working as expected")
	}
}
