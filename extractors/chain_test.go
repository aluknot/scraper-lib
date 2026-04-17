package extractors

import (
	"context"
	"strings"
	"testing"

	"github.com/aluknot/scraper-lib/types"
)

func TestChain_Order(t *testing.T) {
	chain := DefaultChain()

	expected := []string{"domain_specific", "readability", "trafilatura", "fallback"}
	for i, e := range chain.extractors {
		if e.Name() != expected[i] {
			t.Errorf("position %d: expected %s, got %s", i, expected[i], e.Name())
		}
	}
}

func TestChain_FirstSuccess(t *testing.T) {
	// Create chain with a mock extractor that always succeeds
	chain := NewChain(
		&mockExtractor{name: "first", priority: 0, shouldSucceed: true},
		&mockExtractor{name: "second", priority: 1, shouldSucceed: true},
	)

	result, err := chain.Extract(context.Background(), "<html></html>", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExtractorUsed != "first" {
		t.Errorf("expected 'first' extractor, got %s", result.ExtractorUsed)
	}
	// Should have only tried the first extractor
	if len(result.Attempts) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(result.Attempts))
	}
}

func TestChain_FallbackToSecond(t *testing.T) {
	chain := NewChain(
		&mockExtractor{name: "first", priority: 0, shouldSucceed: false},
		&mockExtractor{name: "second", priority: 1, shouldSucceed: true},
	)

	result, err := chain.Extract(context.Background(), "<html></html>", "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ExtractorUsed != "second" {
		t.Errorf("expected 'second' extractor, got %s", result.ExtractorUsed)
	}
	if len(result.Attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(result.Attempts))
	}
}

func TestChain_AllFail(t *testing.T) {
	chain := NewChain(
		&mockExtractor{name: "first", priority: 0, shouldSucceed: false},
		&mockExtractor{name: "second", priority: 1, shouldSucceed: false},
	)

	_, err := chain.Extract(context.Background(), "<html></html>", "https://example.com")
	if err == nil {
		t.Fatal("expected error when all extractors fail")
	}
	if _, ok := err.(ErrAllExtractorsFailed); !ok {
		t.Errorf("expected ErrAllExtractorsFailed, got %T: %v", err, err)
	}
}

func TestDomainSpecificExtractor_NoRule(t *testing.T) {
	e := NewDomainSpecificExtractor()

	// URL that doesn't match any rule
	_, err := e.Extract(context.Background(), "<html></html>", "https://example.com/article")
	if err == nil {
		t.Fatal("expected error for unknown domain")
	}
	if _, ok := err.(ErrNoRuleForDomain); !ok {
		t.Errorf("expected ErrNoRuleForDomain, got %T: %v", err, err)
	}
}

func TestDomainSpecificExtractor_GitHub(t *testing.T) {
	e := NewDomainSpecificExtractor()

	htmlContent := `<html><body>
<article class="markdown-body">
<h1>Test Repository README</h1>
<p>This is a test repository. Lorem ipsum dolor sit amet, consectetur
adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore
magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation
ullamco laboris nisi ut aliquip ex ea commodo consequat.</p>
</article>
</body></html>`

	result, err := e.Extract(context.Background(), htmlContent, "https://github.com/test/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Content, "Test Repository README") {
		t.Errorf("expected title in content, got:\n%s", result.Content)
	}
	if result.ExtractorUsed != "domain_specific:readme" {
		t.Errorf("expected 'domain_specific:readme', got %s", result.ExtractorUsed)
	}
}

func TestValidateRule_Valid(t *testing.T) {
	rule := domainRule{
		Domain:    "example.com",
		PathRegex: "/.*",
		Type:      "article",
		Selector:  "article.content",
	}

	if err := ValidateRule(rule); err != nil {
		t.Errorf("expected valid rule, got error: %v", err)
	}
}

func TestValidateRule_InvalidRegex(t *testing.T) {
	rule := domainRule{
		Domain:    "example.com",
		PathRegex: "[invalid", // Invalid regex
		Type:      "article",
		Selector:  "article.content",
	}

	if err := ValidateRule(rule); err == nil {
		t.Error("expected error for invalid regex")
	}
}

func TestValidateRule_InvalidType(t *testing.T) {
	rule := domainRule{
		Domain:    "example.com",
		PathRegex: "/.*",
		Type:      "invalid_type",
		Selector:  "article.content",
	}

	if err := ValidateRule(rule); err == nil {
		t.Error("expected error for invalid type")
	}
}

// mockExtractor is a test double for the Extractor interface.
type mockExtractor struct {
	name          string
	priority      int
	shouldSucceed bool
}

func (m *mockExtractor) Name() string  { return m.name }
func (m *mockExtractor) Priority() int { return m.priority }
func (m *mockExtractor) Extract(ctx context.Context, htmlContent, url string) (*types.ExtractResult, error) {
	if !m.shouldSucceed {
		return nil, nil // Return nil to simulate low_quality
	}
	return &types.ExtractResult{
		Content:       "Mock content with enough words to be valid for testing purposes. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
		ExtractorUsed: m.name,
		WordCount:     120,
	}, nil
}
