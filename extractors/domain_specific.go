package extractors

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aluknot/scraper-lib/internal/markdown"
	"github.com/aluknot/scraper-lib/internal/urlutil"
	"github.com/aluknot/scraper-lib/types"
	"gopkg.in/yaml.v3"
)

// defaultConfigYAML contains the embedded default domain extraction rules.
const defaultConfigYAML = `
rules:
  - domain: "github.com"
    path_regex: "/.*"
    type: "readme"
    selector: "article.markdown-body"
    alt: "article#readme"
    transform: "markdown"

  - domain: "github.com"
    path_regex: "/[^/]+/[^/]+/releases/.*"
    type: "releases"
    selector: "div.markdown-body"

  - domain: "wikipedia.org"
    path_regex: "/.*"
    type: "article"
    selector: "#mw-content-text"
`

// ErrNoRuleForDomain indicates no domain-specific rule matched.
// This is not a fatal error — the chain continues with the next extractor.
type ErrNoRuleForDomain struct {
	Domain string
	Path   string
}

func (e ErrNoRuleForDomain) Error() string {
	return fmt.Sprintf("no rule for domain %s (path: %s)", e.Domain, e.Path)
}

// ErrSelectorNotFound indicates the CSS selector did not match any element.
type ErrSelectorNotFound struct {
	URL      string
	Selector string
}

func (e ErrSelectorNotFound) Error() string {
	return fmt.Sprintf("selector not found: %s at %s", e.Selector, e.URL)
}

// domainRule defines an extraction rule for a specific domain/path pattern.
type domainRule struct {
	Domain    string `yaml:"domain"`
	PathRegex string `yaml:"path_regex"`
	Type      string `yaml:"type"`
	Selector  string `yaml:"selector"`
	Alt       string `yaml:"alt"`
	Transform string `yaml:"transform"` // "" = HTML, "markdown"
}

type domainConfig struct {
	Rules []domainRule `yaml:"rules"`
}

// DomainSpecificExtractor applies CSS-based extraction rules configured
// externally (YAML file). It knows when it doesn't apply and returns
// ErrNoRuleForDomain so the chain can continue.
type DomainSpecificExtractor struct {
	rules       []domainRule
	compiledREs []*regexp.Regexp
}

// NewDomainSpecificExtractor creates a new extractor, loading rules from
// config/extractors.yaml. Falls back to embedded defaults if the file
// is not found.
func NewDomainSpecificExtractor() *DomainSpecificExtractor {
	rules := loadDefaultRules()

	// Try loading from external config (overrides defaults)
	if external, err := loadExternalRules(""); err == nil {
		rules = external
	}

	// Precompile regex patterns
	compiledREs := make([]*regexp.Regexp, len(rules))
	for i, r := range rules {
		re, err := regexp.Compile(r.PathRegex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scraper-lib: warning: invalid path_regex %q for domain %s: %v\n",
				r.PathRegex, r.Domain, err)
			compiledREs[i] = nil
		} else {
			compiledREs[i] = re
		}
	}

	return &DomainSpecificExtractor{
		rules:       rules,
		compiledREs: compiledREs,
	}
}

func (e *DomainSpecificExtractor) Name() string  { return "domain_specific" }
func (e *DomainSpecificExtractor) Priority() int { return 0 }

func (e *DomainSpecificExtractor) Extract(ctx context.Context, htmlContent, articleURL string) (*types.ExtractResult, error) {
	domain := urlutil.Domain(articleURL)
	path := urlutil.Path(articleURL)

	// Find matching rule
	var rule *domainRule
	for i := range e.rules {
		if e.rules[i].Domain != domain {
			continue
		}
		re := e.compiledREs[i]
		if re == nil {
			continue
		}
		if re.MatchString(path) {
			r := e.rules[i]
			rule = &r
			break
		}
	}

	if rule == nil {
		return nil, ErrNoRuleForDomain{Domain: domain, Path: path}
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parse HTML: %w", err)
	}

	content, err := doc.Find(rule.Selector).Html()
	if err != nil || content == "" {
		if rule.Alt != "" {
			content, err = doc.Find(rule.Alt).Html()
			if err != nil || content == "" {
				return nil, ErrSelectorNotFound{URL: articleURL, Selector: rule.Alt}
			}
		} else {
			return nil, ErrSelectorNotFound{URL: articleURL, Selector: rule.Selector}
		}
	}

	result := &types.ExtractResult{
		Content:       content,
		ExtractorUsed: fmt.Sprintf("domain_specific:%s", rule.Type),
		WordCount:     countWords(content),
	}

	if rule.Transform == "markdown" {
		result.Content = markdown.HTMLToMarkdown(content)
		result.Warnings = append(result.Warnings, "content_transformed_to_markdown")
	}

	return result, nil
}

// loadDefaultRules loads rules embedded in the binary.
func loadDefaultRules() []domainRule {
	var cfg domainConfig
	if err := yaml.Unmarshal([]byte(defaultConfigYAML), &cfg); err != nil {
		return nil
	}
	return cfg.Rules
}

// loadExternalRules loads rules from an external YAML file.
// If path is empty, uses "config/extractors.yaml" relative to cwd.
func loadExternalRules(path string) ([]domainRule, error) {
	if path == "" {
		path = "config/extractors.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg domainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Validate rules at load time
	for _, rule := range cfg.Rules {
		if err := ValidateRule(rule); err != nil {
			return nil, fmt.Errorf("invalid rule for %s: %w", rule.Domain, err)
		}
	}

	return cfg.Rules, nil
}

// ValidateRule validates a rule before use.
// Detects invalid CSS selectors and malformed path_regex patterns.
func ValidateRule(rule domainRule) error {
	// Validate path_regex
	if _, err := regexp.Compile(rule.PathRegex); err != nil {
		return fmt.Errorf("invalid path_regex %q: %w", rule.PathRegex, err)
	}

	// Validate CSS selectors are not empty
	if rule.Selector == "" {
		return fmt.Errorf("selector must not be empty")
	}

	// Validate alt selector if present
	if rule.Alt != "" && strings.TrimSpace(rule.Alt) == "" {
		return fmt.Errorf("alt selector must not be empty if provided")
	}

	// Validate known types
	validTypes := map[string]bool{
		"readme":   true,
		"releases": true,
		"article":  true,
		"product":  true,
		"job":      true,
		"list":     true,
	}
	if !validTypes[rule.Type] {
		return fmt.Errorf("unknown type %q (valid: readme, releases, article, product, job, list)", rule.Type)
	}

	return nil
}
